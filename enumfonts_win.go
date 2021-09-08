//go:build windows
// +build windows

package enumfonts

import (
	"bytes"
	"io/ioutil"
	"sort"
	"strings"
	"syscall"
	"unicode"
	"unsafe"

	"github.com/lxn/win"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func EnumFonts() ([]string, error) {
	type logFont struct {
		LfHeight         int32
		LfWidth          int32
		LfEscapement     int32
		LfOrientation    int32
		LfWeight         int32
		LfItalic         byte
		LfUnderline      byte
		LfStrikeOut      byte
		LfCharSet        byte
		LfOutPrecision   byte
		LfClipPrecision  byte
		LfQuality        byte
		LfPitchAndFamily byte
		LfFaceName       [32]byte
	}
	type logFontEx struct {
		elfLogFont  logFont
		elfFullName [32]byte
		elfStyle    [32]byte
		elfScript   [32]byte
	}

	// load gdi32.dll
	dll, err := syscall.LoadDLL(`gdi32.dll`)
	if err != nil {
		return nil, err
	}

	// load function
	fct := dll.MustFindProc("EnumFontFamiliesExA")

	// create callback function receiving each font name
	fontMap := map[string]bool{}
	var fonts []string
	callback := syscall.NewCallback(func(lpElfe *logFontEx, lpntme int, fontType int, lParam int) (ret uintptr) {
		bts := bytesSlice(lpElfe.elfScript)
		fontStyle := bytes2Str(bts)
		if "Regular" != fontStyle {
			return 1
		}

		bts = bytesSlice(lpElfe.elfLogFont.LfFaceName)
		fontName := bytes2Str(bts)
		if strings.HasPrefix(fontName, "@") || fontMap[fontName] {
			return 1
		}

		if isChinese(fontName) && strings.Contains(fontName, " ") {
			fontName = fontName[:strings.LastIndex(fontName, " ")]
		}
		fontName = strings.TrimSpace(fontName)
		fontMap[fontName] = true
		return 1
	})

	// call function to enumerate font names
	lf := logFont{
		LfCharSet: 1,
	}
	hDC := win.GetDC(0)
	defer win.ReleaseDC(0, hDC)
	_, _, _ = fct.Call(uintptr(hDC), uintptr(unsafe.Pointer(&lf)), callback, 0, 0)

	for f, _ := range fontMap {
		fonts = append(fonts, f)
	}
	sort.Strings(fonts)
	return fonts, nil
}

func bytesSlice(b [32]byte) (ret []byte) {
	ret = make([]byte, 0, len(b))
	for _, c := range b {
		if c == 0 {
			break
		}
		ret = append(ret, c)
	}
	return
}

func bytes2Str(b []byte) string {
	b, err := gbk2utf8(b)
	if nil == err {
		return string(b)
	}
	return string(b)
}

func gbk2utf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func isChinese(str string) bool {
	var count int
	for _, v := range str {
		if unicode.Is(unicode.Han, v) {
			count++
			break
		}
	}
	return count > 0
}
