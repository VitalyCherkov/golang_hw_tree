package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func printPrefix(out io.Writer, prefix []bool, meIsLast bool) error {
	for _, isLastInLevel := range prefix {
		var err error
		if isLastInLevel {
			_, err = fmt.Fprint(out, "\t")
		} else {
			_, err = fmt.Fprint(out, "│\t")
		}
		if err != nil {
			return err
		}
	}

	var err error
	if meIsLast {
		_, err = fmt.Fprint(out, "└───")
	} else {
		_, err = fmt.Fprint(out, "├───")
	}
	return err
}

func printFileName(out io.Writer, file os.FileInfo) error {
	_, err := fmt.Fprint(out, file.Name());
	if err != nil {
		return err
	}

	if !file.IsDir() {
		if file.Size() > 0 {
			_, err = fmt.Fprintf(out, " (%db)", file.Size())
		} else {
			_, err = fmt.Fprint(out, " (empty)")
		}
	}
	return err
}

func printFullFile(out io.Writer, prefix []bool, meIsLast bool, file os.FileInfo) error {
	err := printPrefix(out, prefix, meIsLast)
	if err != nil {
		return err
	}

	err = printFileName(out, file)
	if err != nil {
		return err
	}

	fmt.Fprintln(out)

	return err
}

func getNextIndex(files []os.FileInfo, curIndex int, printFiles bool) (nextIndex int, hasNext bool) {
	if curIndex >= len(files) || len(files) == 0 {
		return nextIndex, false
	}

	nextIndex = curIndex + 1;

	if printFiles {
		if curIndex < len(files) - 1 {
			hasNext = true
		}
		return nextIndex, hasNext
	}

	for ; nextIndex < len(files); nextIndex++ {
		if files[nextIndex].IsDir() {
			return nextIndex, true
		}
	}

	return nextIndex, false
}

func dirTreeRec(out io.Writer, path string, file os.FileInfo, prefix []bool, printFiles bool) error {
	if !file.IsDir() {
		return fmt.Errorf("file [%s] is not dir", file.Name())
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	curIndex, hasNext := getNextIndex(files, -1, printFiles)
	if !hasNext {
		return nil
	}

	for nextIndex, hasNext := getNextIndex(files, curIndex, printFiles);
		curIndex < len(files);
		nextIndex, hasNext = getNextIndex(files, nextIndex, printFiles) {

		nextPath := strings.Join([]string{path, files[curIndex].Name()}, string(os.PathSeparator))

		if files[curIndex].IsDir() || printFiles {
			err := printFullFile(out, prefix, !hasNext, files[curIndex])
			if err != nil {
				return err
			}
		}
		if files[curIndex].IsDir() {
			err := dirTreeRec(out, nextPath, files[curIndex], append(prefix, !hasNext), printFiles)
			if err != nil {
				return err
			}
		}

		curIndex = nextIndex
	}

	return err
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	file, err := os.Stat(path)
	if err != nil {
		return err
	}

	return dirTreeRec(out, path, file, nil, printFiles)
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"

	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}