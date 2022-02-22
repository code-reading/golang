package main

import (
	"errors"
	"fmt"
)

func main() {
	e := Error()
	fmt.Println(e)
	ee := errors.Unwrap(e)
	fmt.Println(ee)
}

func Error() error {
	return fmt.Errorf("this is error, %w", errors.New("Error1"))
}

// fmt.Errorf( %w ) 通过%w  可以嵌入错误， 在通过 errors.Unwrap 还要错误

/*
output
this is error, Error1
Error1
*/
