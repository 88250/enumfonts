// +build !windows

package enumfonts

import "errors"

func EnumFonts() ([]string, error) {
	return nil, errors.New("not implemented on this OS")
}
