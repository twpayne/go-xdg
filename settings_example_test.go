package xdg_test

import (
	"fmt"

	xdg "github.com/twpayne/go-xdg/v3"
)

func ExampleSetting_Check() {
	setting := xdg.Setting{
		Property: xdg.DefaultWebBrowserProperty,
	}
	isGoogleChrome, err := setting.Check("google-chrome.desktop")
	if err != nil {
		panic(err)
	}
	fmt.Println(isGoogleChrome)

}
func ExampleSetting_Get() {
	setting := xdg.Setting{
		Property:    xdg.DefaultURLSchemeHandlerProperty,
		SubProperty: "http",
	}
	value, err := setting.Get()
	if err != nil {
		panic(err)
	}
	fmt.Println(value)
}

func ExampleSetting_Set() {
	setting := xdg.Setting{
		Property:    xdg.DefaultURLSchemeHandlerProperty,
		SubProperty: "http",
	}
	if err := setting.Set("firefox.desktop"); err != nil {
		panic(err)
	}
}
