package model

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"text/template"
)

type Query struct {
	Pn string // 进程名
}

// cpu语句
func (q *Query) CQuery() string {
	var tmplBytes bytes.Buffer
	tmpl := template.New("tpl")
	t, err := tmpl.Parse(`ps --no-headers -o "%cpu" -C {{.}}`)
	if err != nil {
		panic(err)
	}
	_ = t.Execute(&tmplBytes, q.Pn)
	return tmplBytes.String()
}

// 内存语句
func (q *Query) MQuery() string {
	var tmplBytes bytes.Buffer
	tmpl := template.New("tpl")
	t, err := tmpl.Parse(`ps --no-headers -o "rss,cmd" -C {{.}} | awk '{ sum+=$1 } END { printf ("%d%s\n", sum/NR/1024,"M") }'`)
	if err != nil {
		panic(err)
	}
	_ = t.Execute(&tmplBytes, q.Pn)
	return tmplBytes.String()
}

// 采集进程内存占用比
func (q *Query) CollectMem() string {
	d := exec.Command("bash", "-c", q.MQuery())
	out, err := d.CombinedOutput()
	if err != nil {
		fmt.Printf("combined out:\n%s\n", string(out))
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	return string(out)
}

// 采集进程cpu占用比
func (q *Query) CollectCpu() string {
	d := exec.Command("bash", "-c", q.CQuery())
	out, err := d.CombinedOutput()
	if err != nil {
		fmt.Printf("combined out:\n%s\n", string(out))
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}

	return string(out)
}
