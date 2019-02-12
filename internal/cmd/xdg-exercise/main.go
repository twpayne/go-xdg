// Command xdg-exercise exercies various functions and writes their return
// values to stdout in JSON format.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	xdg "github.com/twpayne/go-xdg"
	"github.com/twpayne/go-xdg/xdgsettings"
)

func run() error {
	var result struct {
		BaseDirectorySpecification *xdg.BaseDirectorySpecification
		Settings                   struct {
			DefaultURLSchemeHandler struct {
				HTTP   string
				MailTo string
			}
			WebBrowser struct {
				IsFirefox      bool
				IsGoogleChrome bool
			}
		}
	}

	var err error

	// Exercise xdg.NewBaseDirectorySpecification.
	result.BaseDirectorySpecification, err = xdg.NewBaseDirectorySpecification()
	if err != nil {
		return err
	}

	// Exercise xdgsettings.Get.
	result.Settings.DefaultURLSchemeHandler.HTTP, err = xdgsettings.Get(xdgsettings.DefaultURLSchemeHandlerProperty, "http")
	if err != nil {
		return err
	}
	result.Settings.DefaultURLSchemeHandler.MailTo, err = xdgsettings.Get(xdgsettings.DefaultURLSchemeHandlerProperty, "mailto")
	if err != nil {
		return err
	}

	// Exercise xdgsettings.Check.
	result.Settings.WebBrowser.IsFirefox, err = xdgsettings.Check(xdgsettings.DefaultWebBrowserProperty, "", "firefox.desktop")
	if err != nil {
		return err
	}
	result.Settings.WebBrowser.IsGoogleChrome, err = xdgsettings.Check(xdgsettings.DefaultWebBrowserProperty, "", "google-chrome.desktop")
	if err != nil {
		return err
	}

	// Exercise xdgsettings.Set. This is really bad, but we have no idea what
	// software is installed on the user's machine, so we have no idea what
	// values will not return an error. One might hope that setting
	// property.subProperty to the old value will work, but sadly this not the
	// case: Ubuntu 18.04.1 LTS ships with default-url-scheme-handler mailto set
	// to thunderbird.desktop, but this is not present in the minimal
	// installation and setting the existing value fails. So we fall back to the
	// "http" URL handler, which we hope is a valid value everywhere.
	property, subProperty := xdgsettings.DefaultURLSchemeHandlerProperty, "http"
	oldValue, err := xdgsettings.Get(property, subProperty)
	if err != nil {
		return err
	}
	defer func() {
		xdgsettings.Set(property, subProperty, oldValue)
	}()
	newValue := oldValue
	if err := xdgsettings.Set(property, subProperty, newValue); err != nil {
		return err
	}
	if ok, err := xdgsettings.Check(property, subProperty, newValue); err != nil || !ok {
		return fmt.Errorf("xdgsettings.Check(%q, %q %q) == %v, %v, want true, <nil>", property, subProperty, newValue, ok, err)
	}

	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	return e.Encode(&result)
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
