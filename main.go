package main

import (
	"reflect"
	"fmt"
	"os"
	"strings"
	"strconv"
	"io/ioutil"
)

type ISerializable interface {
	GetFieldNames() []string
	SetFieldValue(field string, value reflect.Value) ISerializable
}

type TestStruct struct {
	Value int `key:"vvalue"`
	TestValue int
	NonSerialize int `Serialize:"false"`
	DoSerialize int `Serialize:"true"`
}

func (this TestStruct) GetFieldNames() (ret []string) {
	ty := reflect.TypeOf(this)
	for i := 0 ; i < ty.NumField(); i++ {
		ret = append(ret, ty.Field(i).Name)
	}
	return
}
func (this TestStruct) SetFieldValue(field string, value reflect.Value) ISerializable {
	tv := reflect.ValueOf(&this)
	s := tv.Elem()
	str := value.String()

	val, _ := strconv.ParseInt(str, 10, 64)
	val = int64(val)

	s.FieldByName(field).SetInt(val)
	return this
}

// 구조체의 정보를 출력함
func PrintInfo(tp interface{}) {
	ty := reflect.TypeOf(tp)	// 구조체의 타입
	tv := reflect.ValueOf(tp)	// 구조체의 값
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println("PANIC : ", rec)
		}
	}()			// 패닉 처리
	if ty.Kind() != reflect.Struct {
		panic("you can print struct's info only")
		return
	}	// 구조체가 아니면 패닉

	// 구조체의 필드 수만큼 반복
	for i := 0; i < ty.NumField(); i++ {
		field := ty.Field(i)

		// Serialize 필드가 false면 출력하지 않는다
		if tag, ok := field.Tag.Lookup("Serialize"); ok {
			if tag == "false" {
				continue
			}
		}

		// key 태그가 있다면 그 값을, 아니라면 필드 이름을 키로 사용
		if tag, ok := field.Tag.Lookup("key"); ok {
			fmt.Println("Key : ", field.Name, "(tag : ", tag, ")\tValue : ", tv.Field(i).Interface())
		} else {
			fmt.Println("Key : ", field.Name, "(tag : ", field.Name, ")\tValue : ", tv.Field(i).Interface())
		}
	}
}

// 구조체를 파일로 저장
func Serialize(path string, st ISerializable) {
	ty := reflect.TypeOf(st)	// 구조체 타입
	tv := reflect.ValueOf(st)	// 구조체 값
	saveStr := ""			// 파일에 저장될 문자열

	// 구조체의 필드 수만큼 반복
	for i := 0; i < ty.NumField(); i++ {
		field := ty.Field(i)

		// Serialize 태그가 false면 저장하지 않는다
		if tag, ok := field.Tag.Lookup("Serialize"); ok {
			if tag == "false" {
				continue
			}
		}

		// key 태그가 있으면 그 값을, 아니면 필드 이름을 키로 쓴다
		if tag, ok := field.Tag.Lookup("key"); ok {
			saveStr += fmt.Sprintln("Key:", tag, ",Value:", tv.Field(i).Interface())
		} else {
			saveStr += fmt.Sprintln("Key:", field.Name, ",Value:", tv.Field(i).Interface())
		}
	}

	// 파일에 문자열을 기록
	fp, _ := os.Create(path + ".serial")
	fp.WriteString(saveStr)
	fp.Close()
}

// 파일로부터 구조체 로드
func Deserialize(path string, st ISerializable) ISerializable {

	fields := st.GetFieldNames()			// 구조체의 필드 이름들
	b, _ := ioutil.ReadFile(path + ".serial")	// 파일의 텍스트를 바이트 배열로 받는다
	rawData := strings.Split(string(b), "\n")	// 바이트 배열을 문자열로 변환하고 줄 별로 나눈다
	loopCnt := 0					// 구조체에 값을 저장한 횟수

	for _, str := range rawData {
		if len(str) <= 0 {
			continue
		}
		// 문자열에서 키랑 값을 분리
		pair := strings.Split(str, " ,")
		key := pair[0][5:]
		val := pair[1][7:]

		// 키랑 같은 이름의 key 태그를 가지고 있으면 그 필드의 이름을 키로 사용
		for _, field := range fields {
			f, _ := reflect.TypeOf(st).FieldByName(field)
			if v, ok := f.Tag.Lookup("key"); ok {
				if v == key {
					key = field
				}
			}
		}

		// 구조체에 값을 저장
		st = st.SetFieldValue(key, reflect.ValueOf(val))
		loopCnt++
	}

	// 파일이 비어있었다면 new로 새로 만들어서 기본값을 채워준다
	if loopCnt == 0 {
		n := reflect.New(reflect.TypeOf(st))
		st = n.Elem().Interface().(ISerializable)
	}
	return st
}

func main() {
	var tp TestStruct
	tp = Deserialize("TestPair", tp).(TestStruct)

	defer func() {
		Serialize("TestPair", tp)
	}()

	PrintInfo(tp)
}