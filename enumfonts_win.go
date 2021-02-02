// +build windows

package enumfonts

import (
	"github.com/lxn/win"
	"strings"
	"syscall"
	"unsafe"
)

func EnumFonts() ([]string, error) {
	const lf_FACESIZE = 32
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
		LfFaceName       [lf_FACESIZE]uint8
	}
	type logFontEx struct {
		elfLogFont  logFont
		elfFullName [lf_FACESIZE]uint16
		elfStyle    [lf_FACESIZE]uint16
		elfScript   [lf_FACESIZE]uint16
	}

	// load gdi32.dll
	dll, err := syscall.LoadDLL(`gdi32.dll`)
	if err != nil {
		return nil, err
	}

	// load function
	fct := dll.MustFindProc("EnumFontFamiliesExA")

	// create callback function receiving each font name
	var fonts []string
	callback := syscall.NewCallback(func(lpElfe *logFontEx, lpntme int, fontType int, lParam int) (ret uintptr) {
		bts := make([]byte, 0, len(lpElfe.elfLogFont.LfFaceName))
		for _, c := range lpElfe.elfLogFont.LfFaceName {
			if c == 0 {
				break
			}
			bts = append(bts, c)
		}
		fontName := string(bts)
		if !strings.HasPrefix(fontName, "@") {
			fonts = append(fonts, fontName)
		}
		return 1
	})

	// call function to enumerate font names
	lf := logFont{
		LfCharSet: 1,
	}
	hDC := win.GetDC(0)
	defer win.ReleaseDC(0, hDC)
	_, _, _ = fct.Call(uintptr(hDC), uintptr(unsafe.Pointer(&lf)), callback, 0, 0)

	return fonts, nil
}
