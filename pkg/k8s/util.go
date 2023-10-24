/*
 * @FilePath: /Users/zengshenglong/Code/GoWorkSpace/operators/es-operator/pkg/k8s/util.go
 * @Author: kbsonlong kbsonlong@gmail.com
 * @Date: 2023-10-10 11:23:38
 * @LastEditors: kbsonlong kbsonlong@gmail.com
 * @LastEditTime: 2023-10-24 15:59:59
 * @Description:
 * Copyright (c) 2023 by kbsonlong, All Rights Reserved.
 */
package k8s

import (
	"bytes"
	"fmt"
	"text/template"
)

func ParseConf(conftemp string, es map[interface{}]interface{}) bytes.Buffer {
	funcMap := template.FuncMap{
		"loop": func(name string, to int32) <-chan string {
			ch := make(chan string)
			go func() {
				for i := 0; i <= int(to)-1; i++ {
					ch <- fmt.Sprintf("%s-%d", name, i)
				}
				close(ch)
			}()
			return ch
		},
	}

	tmpl, err := template.New("conf").Funcs(funcMap).Parse(conftemp)
	if err != nil {
		fmt.Println(err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, es)
	if err != nil {
		panic(err)
	}
	if err != nil {
		fmt.Println(err)
	}
	return buf
}
