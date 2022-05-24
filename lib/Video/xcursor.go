package Video

/*
#include <stdio.h>
#include <stdint.h>
#include <stdlib.h>
#include <X11/Xlib.h>
#include <X11/extensions/Xfixes.h>
#cgo LDFLAGS: -L. -lX11 -lXfixes

static Display *display = NULL;

typedef struct
{
	int x, y, width, height;
	uint32_t cursor_serial;
	uint32_t *pixels;
} CustomCursor;

CustomCursor cursor;

int init(const char *displayname)
 {
	XInitThreads();
	if ((display = XOpenDisplay(displayname)) == NULL)
	{
		printf("cannot open display \"%s\"\n", displayname ? displayname : "DEFAULT");
		return -1;
	}
}

CustomCursor *get_mouse_pointer()
{
	XLockDisplay(display);
	XFixesCursorImage *x11_cursor = XFixesGetCursorImage(display);
	XUnlockDisplay(display);

	int x, y;
	uint32_t *pixel32 = (uint32_t*) malloc(x11_cursor->width * x11_cursor->height * sizeof(uint32_t));
	for (y = 0; y < x11_cursor->height; y++)
	{
		for (x = 0; x < x11_cursor->width; x++)
		{
			uint32_t ofs = x + y * x11_cursor->width;
			*(pixel32 + ofs) = *(x11_cursor->pixels + ofs);
		}
	}

	cursor.x = x11_cursor->x;
	cursor.y = x11_cursor->y;
	cursor.width = (int) x11_cursor->width;
	cursor.height = (int) x11_cursor->height;
	cursor.pixels = pixel32;
	cursor.cursor_serial = x11_cursor->cursor_serial;

	return &cursor;
}


*/
import "C"

import (
	"os"
	"unsafe"
)

func CursorInit() {
	display := os.Getenv("DISPLAY")
	C.init(C.CString(display))

}

func CursorGet() (int, int, []byte, int, int, int) {
	cursorPointer := C.get_mouse_pointer()
	cursor := *cursorPointer
	pixels := unsafe.Pointer(cursor.pixels)
	data := C.GoBytes(pixels, C.int(cursor.width)*C.int(cursor.height)*4)
	x := int(C.int(cursor.x))
	y := int(C.int(cursor.y))
	width := int(C.int(cursor.width))
	height := int(C.int(cursor.height))
	serial := int(C.int(cursor.cursor_serial))

	return x, y, data, width, height, serial

}
