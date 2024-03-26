// dbl
package main

import (
	"archive/zip"
	"bufio"
	"compress/flate"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

func main() {
	cfg, err := ini.Load("dbl.ini")
	if err != nil {
		fmt.Printf("Ошибка чтения ini файла: %v\n", err)
		os.Exit(1)
	}
	s := cfg.Section("dbl")
	SourceFolder := s.Key("SourceFolder").String()
	Sources := s.Key("Sources").Strings(",")
	DstFolders := s.Key("DstFolders").Strings(",")
	CompressLevel := s.Key("CompressLevel").RangeInt(-1, -2, 9)

	test := true
	for _, source := range Sources {
		if f, err := os.Open(sum(SourceFolder, source)); err != nil {
			log.Println("Нет доступа к файлу ", sum(SourceFolder, source))
			log.Println(err)
			test = false
		} else {
			f.Close()
		}
	}
	if test {
		t := time.Now()
		name := fmt.Sprintf("%d_%02d_%02d_%02d%02d%02d.zip",
			t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())

		backup := func() {
			defer func() {
				if err := recover(); err != nil {
					log.Println("Архивация прервана. Ошибка:", err)
				}
			}()

			log.Println("Архивация начата")
			wzip := make([]io.WriteCloser, len(DstFolders))
			for i, dst := range DstFolders {
				wzip[i], err = os.Create(sum(dst, name))
				if err != nil {
					panic(err)
				}
			}
			fz := MultiWriter(wzip...)
			defer fz.Close()
			fbuf := bufio.NewWriterSize(fz, 2*1024*1024)
			defer fbuf.Flush()
			w := zip.NewWriter(fbuf)
			compressor := func(out io.Writer) (io.WriteCloser, error) {
				return flate.NewWriter(out, CompressLevel)
			}
			w.RegisterCompressor(zip.Deflate, compressor)
			defer w.Close()
			for _, source := range Sources {
				f, err := os.Open(sum(SourceFolder, source))
				if err != nil {
					panic(err)
				}
				zf, err := w.Create(source)
				if err != nil {
					panic(err)
				}
				_, err = io.Copy(zf, f)
				if err != nil {
					panic(err)
				}
				f.Close()
			}
			log.Println("Архивация успешно завершена")
		}

		backup()
	} else {
		log.Println("Архивация не выполнена")
	}
	log.Println("Программа завершается")
	fmt.Println("\n=========================\n")
	fmt.Println("Нажмите Enter для выхода.")
	fmt.Scanln()
}

func sum(s1, s2 string) string { return strings.TrimRight(s1, "/\\") + "/" + s2 }
