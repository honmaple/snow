/*********************************************************************************
 Copyright © 2019 jianglin
 File Name: common.go
 Author: jianglin
 Email: mail@honmaple.com
 Created: 2019-04-04 14:24:22 (CST)
 Last Update: Saturday 2019-09-07 16:41:44 (CST)
		  By:
 Description:
 *********************************************************************************/
package core

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

// DateFormats ..
var DateFormats = []string{"2006-01-02 15:04:05", "2006-01-02 15:04", "2006-01-02"}

// parseTime ..
func ParseTime(value string) (time.Time, error) {
	for _, f := range DateFormats {
		if date, err := time.Parse(f, value); err == nil {
			return date, nil
		}
	}
	return time.Time{}, fmt.Errorf("date format error")
}

func SetField(obj interface{}, name string, value interface{}) error {
	field := reflect.ValueOf(obj).Elem().FieldByName(name)

	if !field.IsValid() {
		return fmt.Errorf("No such field: %s in %T", name, obj)
	}

	if !field.CanSet() {
		return fmt.Errorf("Cannot set %s field value", name)
	}
	fieldType := field.Type()
	val := reflect.ValueOf(value)
	if fieldType != val.Type() {
		return fmt.Errorf("Provided value type didn't match obj field type")
	}
	field.Set(val)
	return nil
}

func GetField(obj interface{}, name string) (interface{}, error) {
	var value reflect.Value

	kind := reflect.TypeOf(obj).Kind()
	if kind != reflect.Struct && kind != reflect.Ptr {
		return nil, fmt.Errorf("Cannot use GetField on a non-struct interface")
	}
	if kind == reflect.Ptr {
		value = reflect.ValueOf(obj).Elem()
	} else {
		value = reflect.ValueOf(obj)
	}
	field := value.FieldByName(name)
	if !field.IsValid() {
		return nil, fmt.Errorf("No such field: %s in %T", name, obj)
	}
	return field.Interface(), nil
}

func CopyFile(src, dst string) (written int64, err error) {
	if stat, err := os.Stat(src); err != nil {
		return 0, err
	} else if !stat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return
	}
	defer srcFile.Close()
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer dstFile.Close()
	return io.Copy(dstFile, srcFile)
}

func CopyDir(src, dst string) (written int64, err error) {
	files, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}
	for _, file := range files {
		srcFile := filepath.Join(src, file.Name())
		dstFile := filepath.Join(dst, file.Name())
		if !file.IsDir() {
			return CopyDir(srcFile, dstFile)
		}
		if written, err = CopyFile(srcFile, dstFile); err != nil {
			return
		}
	}
	return
}

func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsExist(err) || err == nil {
		return true
	}
	return false
}

func ListFiles(path string) ([]string, error) {
	files := make([]string, 0)
	rd, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, f := range rd {
		filename := filepath.Join(path, f.Name())
		if !f.IsDir() {
			files = append(files, filename)
			continue
		}
		fs, err := ListFiles(filename)
		if err != nil {
			return nil, err
		}
		files = append(files, fs...)
	}
	return files, nil
}

func StringReplace(s string, vars map[string]string) string {
	if vars == nil || s == "" {
		return s
	}
	args := make([]string, 0)
	for k, v := range vars {
		args = append(args, k)
		args = append(args, v)
	}
	r := strings.NewReplacer(args...)
	return r.Replace(s)

}

// CheckInList ..
func CheckInList(f []string, v string) bool {
	for i := range f {
		if f[i] == v {
			return true
		}
	}
	return false
}

func PrettyPrint(i interface{}) {
	s, _ := json.MarshalIndent(i, "", "  ")
	fmt.Println(string(s))
}

func Exit(message string, code int) {
	fmt.Println(message)
	os.Exit(code)
}
