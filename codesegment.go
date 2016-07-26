package main

import (
    "bufio"
    "crypto/sha1"
    "encoding/base64"
    "errors"
    "fmt"
    "os"
    "strings"
)

const IdLen = 28

type CodeSegment struct {
    Id, Category, Tags, Desc, Code string
}

func (cs CodeSegment) PrintToScreen() {
    fmt.Printf("ID:%s CATEGORY:%s TAGS:%s\n", cs.Id, cs.Category, cs.Tags)
    fmt.Printf("DESCRIPTION: %s\n", cs.Desc)
    codeLines := strings.Split(cs.Code, "\n")
    for i, line := range codeLines {
        if i == 0 {
            fmt.Println("CONTENT:", line)
        } else {
            fmt.Println("        ", line)
        }
    }
}

var CodePrefixSpace string = "          "  // len:10
func (cs CodeSegment) PrintToFile(fpath string) error {
    f, err := os.OpenFile(fpath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0660)
    if err != nil {
        return err
    }
    defer f.Close()

    f.WriteString("Id:       " + cs.Id + "\n")
    f.WriteString("Category: " + cs.Category + "\n")
    f.WriteString("Tags:     " + cs.Tags + "\n")
    f.WriteString("Desc:     " + cs.Desc + "\n")

    codeLines := strings.Split(cs.Code, "\n")
    for i, line := range codeLines {
        if i == 0 {
            f.WriteString("Content:  " + line + "\n")
        } else {
            f.WriteString(CodePrefixSpace + line + "\n")
        }
    }
    return nil
}

func (cs *CodeSegment) ReadFromFile(fpath string) error {
    f, err := os.Open(fpath)
    if err != nil {
        return err
    }
    defer f.Close()

    isCodeLine := false
    scanner := bufio.NewScanner(f)
    //var codePrefixSpace string
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, "Id:") {
            cs.Id = strings.TrimSpace(line[len("Id:"):])
        } else if strings.HasPrefix(line, "Category:") {
            cs.Category = strings.TrimSpace(line[len("Category:"):])
        } else if strings.HasPrefix(line, "Tags:") {
            cs.Tags = strings.TrimSpace(line[len("Tags:"):])
        } else if strings.HasPrefix(line, "Desc:") {
            cs.Desc = strings.TrimSpace(line[len("Desc:"):])
        } else if strings.HasPrefix(line, "Content:") {
            cs.Code = strings.TrimSpace(line[len("Content:"):])
            isCodeLine = true
        } else {
            if isCodeLine {
                var codeLine string
                if strings.HasPrefix(line, CodePrefixSpace) {
                    codeLine = line[len(CodePrefixSpace):]
                } else {
                    codeLine = strings.TrimSpace(line)
                }
                cs.Code = cs.Code + "\n" + codeLine
            }
        }
    }
    return nil
}

type Store interface {
    Add(cs CodeSegment) error
    Update(cs CodeSegment) error
    Append(id string, extraContent string) error
    Search(category string, tagStr string) []CodeSegment
    Remove(id string) error
    GetById(id string) (CodeSegment, error)
}

type FileStore struct {
    FilePath string
}

func (fs *FileStore) codeSegmentToStr(cs CodeSegment) string {
    descB64 := base64.StdEncoding.EncodeToString([]byte(cs.Desc))
    contentB64 := base64.StdEncoding.EncodeToString([]byte(cs.Code))
    return cs.Id + "|" + cs.Category + "|" + cs.Tags + "|" + descB64 + "|" + contentB64
}

func (fs *FileStore) strToCodeSegment(str string) (cs CodeSegment, err error) {
    flds := strings.Split(str, "|")
    if len(flds) != 5 {
        err = errors.New("parse segemnt str failed: " + str)
        return
    }
    id := flds[0]
    cate := flds[1]
    tags := flds[2]
    desc := flds[3]
    code := flds[4]

    descBs, err := base64.StdEncoding.DecodeString(desc)
    if err != nil {
        return
    }

    codeBs, err := base64.StdEncoding.DecodeString(code)
    if err != nil {
        return
    }

    desc = string(descBs)
    code = string(codeBs)

    cs = CodeSegment{id, cate, tags, desc, code}
    return
}

func (fs *FileStore) genId(cs CodeSegment) (id string, err error) {
    if cs.Id != "" {
        err = errors.New("id already exists")
        return
    }

    if cs.Category == "" || cs.Tags == "" {
        err = errors.New("generate id failed: category or tags is empty")
    }

    idBytes := sha1.Sum([]byte(cs.Category + cs.Tags))
    id = base64.StdEncoding.EncodeToString(idBytes[:])
    return
}

func (fs *FileStore) Add(cs CodeSegment) error {
    if cs.Id == "" {
        id, err := fs.genId(cs)
        if err != nil {
            return err
        }
        //fmt.Println("id: ", id, "id len:", len(id))
        cs.Id = id
    }

    if err := fs.isDuplicate(cs); err != nil {
        return err
    }

    f, err := os.OpenFile(fs.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0660)
    if err != nil {
        return err
    }
    defer f.Close()

    line := fs.codeSegmentToStr(cs)
    _, err = f.WriteString(line + "\n")
    return err
}

func (fs *FileStore) GetById(id string) (cs CodeSegment, err error) {
    if len(id) < IdLen {
        err = errors.New("invalid id:" + id)
        return
    }

    f, err := os.Open(fs.FilePath)
    if err != nil {
        return
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, id) {
            return fs.strToCodeSegment(line)
        }
    }

    err = errors.New("can not find code-segment by id:" + id)
    return
}

func (fs *FileStore) Update(cs CodeSegment) error {
    newCs, err := fs.GetById(cs.Id)
    if err != nil {
        return err
    }

    if cs.Category != "" {
        newCs.Category = cs.Category
    }

    if cs.Tags != "" {
        newCs.Tags = cs.Tags
    }

    if cs.Desc != "" {
        newCs.Desc = cs.Desc
    }

    if cs.Code != "" {
        newCs.Code = cs.Code
    }

    fs.Remove(cs.Id)
    return fs.Add(newCs)
}

func (fs *FileStore) Append(id string, extraContent string) error {
    newCs, err := fs.GetById(id)
    if err != nil {
        return err
    }

    newCs.Code = strings.Trim(newCs.Code, "\n") + "\n" + strings.Trim(extraContent, "\n")
    fs.Remove(id)
    return fs.Add(newCs)
}

func (fs *FileStore) Search(category string, tagStr string) []CodeSegment {
    tags := strings.Split(tagStr, ",")
    matchedLines := grepFile(fs.FilePath, category, tags)
    matchedCs := []CodeSegment{}
    for _, line := range matchedLines {
        cs, err := fs.strToCodeSegment(line)
        if err != nil {
            fmt.Println(err.Error())
        }
        matchedCs = append(matchedCs, cs)
    }
    return matchedCs
}

func grepFile(file string, cate string, tags []string) []string {
    var categoryMatch = func(cateInStore string, c string) bool {
        if c == "" {
            return true
        }
        return strings.HasPrefix(cateInStore, c)
    }

    var tagsMatch = func(tagsInStore string, ss []string) bool {
        if ss == nil || len(ss) == 0 {
            return true
        }
        for _, s := range ss {
            if !strings.Contains(tagsInStore, strings.ToUpper(s)) { // TODO: improve this logic .
                return false
            }
        }
        return true
    }

    cate = strings.ToUpper(cate)
    res := []string{}
    f, err := os.Open(file)
    if err != nil {
        fmt.Println(err)
        return res
    }
    defer f.Close()
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        line := scanner.Text()
        flds := strings.Split(line, "|")
        category := strings.ToUpper(flds[1])
        tagStr := strings.ToUpper(flds[2])
        tagStr = category + "," + tagStr
        if categoryMatch(category, cate) && tagsMatch(tagStr, tags) {
            res = append(res, line)
        }
    }
    if err := scanner.Err(); err != nil {
        fmt.Fprintln(os.Stderr, err)
    }
    return res
}

func (fs *FileStore) Remove(id string) error {
    if len(id) < IdLen {
        return errors.New("Invalid id, id is too short")
    }

    f, err := os.Open(fs.FilePath)
    if err != nil {
        return err
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    fLines := []string{}
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, id) {
            continue
        }
        fLines = append(fLines, line)
    }

    oldFilePath := fs.FilePath + ".old"
    os.Remove(oldFilePath)
    err = os.Rename(fs.FilePath, oldFilePath)
    if err != nil {
        fmt.Println(err.Error())
    }

    newFile, err := os.OpenFile(fs.FilePath, os.O_CREATE|os.O_WRONLY, 0660)
    if err != nil {
        return err
    }
    defer newFile.Close()
    for _, line := range fLines {
        newFile.WriteString(line + "\n")
    }

    return nil
}

func (fs *FileStore) isDuplicate(cs CodeSegment) error {
    //codeB64 := base64.StdEncoding.EncodeToString([]byte(cs.Code))
    f, err := os.Open(fs.FilePath)
    if err != nil {
        fmt.Println(err)
        return nil
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        // bsLine := scanner.Bytes()
        // if bytes.Contains(bsLine, []byte(cs.Id)) || bytes.Contains(bsLine, []byte(codeB64)) {
        //     return
        // }
        line := scanner.Text()
        csInFile, _ := fs.strToCodeSegment(line)
        if csInFile.Code == cs.Code {
            return errors.New("duplicated code content with id:" + csInFile.Id)
        }
        if csInFile.Id == cs.Id {
            return errors.New("duplicated id generated. category and tags is the same with code segment " + csInFile.Id)
        }
    }
    return nil
}
