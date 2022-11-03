package main

import (
	"fmt"
	"github.com/robfig/cron"
	"github.com/spf13/viper"
	"github.com/xuri/excelize/v2"
	"log"
	"os"
	"path/filepath"
	"report/model"
	"sync"
	"time"
)

type hostInfo struct {
	Pn      string `excel:"column:B;desc:进程名;width:30"` // 进程名
	memInfo string `excel:"column:C;desc:内存占用;width:30"`
	cpuInfo string `excel:"column:D;desc:CPU占用比;width:30"`
	Time    string `excel:"column:E;desc:采集时间;width:30"`
}

type server struct {
	Process []string
	Path    string
	Cron    string
}

type config struct {
	Server server
}

var wg sync.WaitGroup

func WriteByStream(hosts *hostInfo, path string) error {
	csv := filepath.Join(path, "/process.xlsx")

	_, err := os.Stat(csv)
	if os.IsNotExist(err) {
		s := excelize.NewFile()
		index := s.NewSheet("Sheet1")
		s.SetCellValue("Sheet1", "A1", "序号")
		s.SetCellValue("Sheet1", "B1", "进程名")
		s.SetCellValue("Sheet1", "C1", "内存占用")
		s.SetCellValue("Sheet1", "D1", "CPU占用比")
		s.SetCellValue("Sheet1", "E1", "采集时间")
		s.SetActiveSheet(index)
		if err := s.SaveAs(csv); err != nil {
			fmt.Println(err)
		}
	}

	f, _ := excelize.OpenFile(csv)

	rows, err := f.GetRows("Sheet1") // len(rows) - 行数
	if err != nil {
		return fmt.Errorf("获取行内容失败: %s", err.Error())
	}

	sw, err := f.NewStreamWriter("Sheet1")
	if err != nil {
		return fmt.Errorf("创建写入流失败: %s", err.Error())
	}

	cols, _ := f.GetCols("Sheet1")
	if err != nil {
		return fmt.Errorf("获取列内容失败: %s", err.Error())
	}

	// 写入之前的数据
	for index, row := range rows {
		rowP := make([]interface{}, len(cols))
		for coid := 0; coid < len(cols); coid++ {
			if row == nil {
				rowP[coid] = nil
			} else {
				rowP[coid] = row[coid]
			}
		}
		cell, _ := excelize.CoordinatesToCellName(1, index+1)
		if err := sw.SetRow(cell, rowP); err != nil {
			return fmt.Errorf("写入原内容失败: %s", err.Error())
		}
	}

	for i := len(rows) + 1; i < len(rows)+2; i++ {
		axis := fmt.Sprintf("A%d", i) // 3 + 1 = 4
		err := sw.SetRow(axis, []interface{}{i - 1, hosts.Pn, hosts.memInfo, hosts.cpuInfo, hosts.Time})
		if err != nil {
			return fmt.Errorf("追加写入失败: %s", err.Error())
		}
	}

	if err := sw.Flush(); err != nil {
		return fmt.Errorf("刷新失败: %s", err.Error())
	}

	if err := f.SaveAs(csv); err != nil {
		return fmt.Errorf("保存失败: %s", err.Error())
	}
	return nil
}

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("read config failed,err:", err)
	}
	var c config
	err := viper.Unmarshal(&c)
	if err != nil {
		fmt.Println(err.Error())
	}

	s := cron.New()
	spec := c.Server.Cron
	err = s.AddFunc(spec, func() {
		for _, v := range c.Server.Process {
			q := model.Query{Pn: v}
			h := &hostInfo{
				Pn:      v,
				memInfo: q.CollectMem(),
				cpuInfo: q.CollectCpu(),
				Time:    time.Now().Format("2006-01-02 15:04:05"),
			}
			err := WriteByStream(h, c.Server.Path)
			if err != nil {
				fmt.Println(err)
				break
			}
		}
	})
	s.Start()

	if err != nil {
		fmt.Println(err)
		s.Stop()
	}

	select {}
}

// TODO
// 1.24小时 停一次 并将数据备份
