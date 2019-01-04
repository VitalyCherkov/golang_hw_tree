package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

//type FileTree struct {
//	children []FileTree
//	info os.FileInfo
//}

//// Добавляет разделенный по кускам путь в дерево
//func (ft *FileTree) appendSpread(pathChunks []string, info os.FileInfo) error {
//	if len(pathChunks) == 0 {
//		return fmt.Errorf("incorrect path: %s", strings.Join(pathChunks, "/"))
//	}
//
//	currentPos := -1
//	for index, val := range ft.children {
//		if val.info.Name() == pathChunks[0] {
//			currentPos = index
//		}
//	}
//
//	if currentPos == -1 {
//		if len(pathChunks) == 1 {
//			ft.children = append(ft.children, FileTree{
//				info:info,
//			})
//			return nil
//		} else {
//			return fmt.Errorf("incorrect path: %s", strings.Join(pathChunks, "/"))
//		}
//	}
//
//	return ft.children[currentPos].appendSpread(pathChunks[1:], info)
//}
//
//// Разделяет путь и удаляет из него начальную часть, метод выше
//func (ft *FileTree) Append(entryPath string, path string, info os.FileInfo) error {
//	entryPathChunks := strings.Split(entryPath, string(filepath.Separator))
//	pathChunks := strings.Split(path, string(filepath.Separator))
//
//	if len(pathChunks) == 0 || len(entryPathChunks) == 0 {
//		return errors.New("empty string")
//	}
//
//	if entryPathChunks[0] == "." {
//		entryPathChunks = entryPathChunks[1:]
//	}
//	if pathChunks[0] == "." {
//		pathChunks = pathChunks[1:]
//	}
//
//	pathChunks = pathChunks[len(entryPathChunks):]
//
//	if len(pathChunks) > 0 {
//		return ft.appendSpread(pathChunks, info)
//	}
//	return nil
//}
//
//// Формирует имя файла
//func (ft *FileTree) getFileName() string {
//	if ft.info == nil {
//		return ""
//	}
//
//	if ft.info.IsDir() {
//		return ft.info.Name()
//	}
//
//	if ft.info.Size() > 0 {
//		return fmt.Sprintf("%s (%db)", ft.info.Name(), ft.info.Size())
//	}
//	return fmt.Sprintf("%s (empty)", ft.info.Name())
//}
//
//// Преробразует дерево к массиву строк
//func (ft *FileTree) toTreeStrings(prefix string, isLast bool) []string {
//	ownPrefix := "├───"
//	if isLast {
//		ownPrefix = "└───"
//	}
//
//	var result = []string{prefix + ownPrefix + ft.getFileName()}
//
//	childPrefix := "│\t"
//	if isLast {
//		childPrefix = "\t"
//	}
//
//	for index, child := range ft.children {
//		result = append(result, child.toTreeStrings(prefix + childPrefix, index == len(ft.children) - 1)...)
//	}
//
//	return result
//}
//
//// Формирует вывод
//func (ft *FileTree) String() string {
//	var result []string
//	for index, subtree := range ft.children {
//		result = append(result, subtree.toTreeStrings("", index == len(ft.children) - 1)...)
//	}
//
//	return strings.Join(result, "\n")
//}

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
	var err error
	_, err = fmt.Fprint(out, file.Name());
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
	var err error

	err = printPrefix(out, prefix, meIsLast)
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
		hasNext = false
		return
	}

	nextIndex = curIndex + 1;

	if printFiles {
		if curIndex < len(files) - 1 {
			hasNext = true
		}
		return
	}

	for ; nextIndex < len(files); nextIndex++ {
		if files[nextIndex].IsDir() {
			hasNext = true
			return
		}
	}

	nextIndex = len(files)
	hasNext = false
	return
}

func dirTreeRec(out io.Writer, path string, file os.FileInfo, prefix []bool, meIsLast bool, printFiles bool, printMe bool) error {
	if !file.IsDir() && !printFiles {
		return fmt.Errorf("file [%s] is not dir", file.Name())
	}

	var err error

	if printMe {
		err = printFullFile(out, prefix, meIsLast, file)
		if err != nil {
			return err
		}
	}

	if !file.IsDir() {
		return nil
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
		nextPrefix := prefix
		if printMe {
			nextPrefix = append(prefix, meIsLast)
		}

		err = dirTreeRec(out, nextPath, files[curIndex], nextPrefix, !hasNext, printFiles, true)
		if err != nil {
			return err
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

	return dirTreeRec(out, path, file, nil, false, printFiles, false)
}

//func dirTree2(out io.Writer, path string, printFiles bool) error {
//
//	var ft FileTree
//
//	err := filepath.Walk(path, func(curPath string, info os.FileInfo, err error) error {
//		if err != nil {
//			return err
//		}
//
//		if info.IsDir() || printFiles {
//			return ft.Append(path, curPath, info)
//		}
//		return nil
//	})
//
//	if err != nil {
//		return err
//	}
//
//	fmt.Fprintln(out, ft.String())
//
//	return nil
//}

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
