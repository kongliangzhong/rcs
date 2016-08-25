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

const IdLen = 27

type CodeSegment struct {
    Id, Category, Tags, Desc, Code string
}

func (cs CodeSegment) PrintToScreen() {
    fmt.Printf("  ID: %s\nCATE: %s\nTAGS: %s\n", cs.Id, cs.Category, cs.Tags)
    fmt.Printf("DESC: %s\n", cs.Desc)
    codeLines := strings.Split(cs.Code, "\n")
    for i, line := range codeLines {
        if i == 0 {
            fmt.Println("CONTENT:")
            fmt.Println("      " + line)
        } else {
            fmt.Println("      " + line)
        }
    }
}

var CodePrefixSpace string = "          " // len:10
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
    isDescLine := false
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
            isDescLine = true
            isCodeLine = false
        } else if strings.HasPrefix(line, "Content:") {
            cs.Code = strings.TrimSpace(line[len("Content:"):])
            isCodeLine = true
            isDescLine = false
        } else {
            if isDescLine {
                var descLine string
                if strings.HasPrefix(line, CodePrefixSpace) {
                    descLine = line[len(CodePrefixSpace):]
                } else {
                    descLine = strings.TrimSpace(line)
                }
                cs.Desc = cs.Desc + "\n" + descLine
                //fmt.Println("cs.Desc:", cs.Desc)
            }

            if isCodeLine {
                var codeLine string
                if strings.HasPrefix(line, CodePrefixSpace) {
                    codeLine = line[len(CodePrefixSpace):]
                } else {
                    codeLine = strings.TrimSpace(line)
                }
                cs.Code = cs.Code + "\n" + codeLine
                //fmt.Println("cs.Code:", cs.Code)
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
    GetStats() RcsStats
}

type RcsStats struct {
    TotalRcsSize int
    AllCates     []string
    AllTags      []string
    CateTagsMap  map[string][]string
    CateNumMap   map[string]int
    TagCatesMap  map[string][]string
    TagNumMap    map[string]int
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

    if cs.Category == "" && cs.Tags == "" {
        err = errors.New("generate id failed: category and tags both empty")
    }

    idBytes := sha1.Sum([]byte(cs.Category + cs.Tags))
    id = base64.StdEncoding.EncodeToString(idBytes[:])
    id = id[:len(id)-1] //
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
    //tags := strings.Split(tagStr, ",")
    matchedLines := grepFile(fs.FilePath, category, tagStr)
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

func grepFile(file string, reqCate string, reqTagStr string) []string {
    var categoryMatch = func(cateInStore string, cateReq string) bool {
        if cateReq == "" {
            return true
        }

        cateInStore = strings.ToUpper(cateInStore)
        cateReq = strings.ToUpper(cateReq)

        cates := strings.Split(cateInStore, "-")
        for _, cate := range cates {
            if cate == cateReq {
                return true
            }
        }
        return false
    }

    var tagsMatch = func(tagsInStore string, reqTagStr string) bool {
        if reqTagStr == "" {
            return true
        }

        tagsInStore = strings.ToUpper(tagsInStore)
        reqTagStr = strings.ToUpper(reqTagStr)

        allTagsOfCs := strings.Split(tagsInStore, ",")
        for _, t := range allTagsOfCs {
            subTs := strings.Split(t, "-")
            if len(subTs) > 1 {
                for _, subTag := range subTs {
                    allTagsOfCs = append(allTagsOfCs, subTag)
                }
            }
        }

        reqTags := strings.Split(reqTagStr, ",")

        for _, reqTag := range reqTags {
            isContains := false
            for _, tagOfCs := range allTagsOfCs {
                if tagOfCs == reqTag {
                    isContains = true
                }
            }

            if !isContains {
                return false
            }
        }
        return true
    }

    //cate = strings.ToUpper(cate)
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
        cateInStore := flds[1]
        tagStr := flds[2]
        tagStr = cateInStore + "," + tagStr
        if categoryMatch(cateInStore, reqCate) && tagsMatch(tagStr, reqTagStr) {
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
    f, err := os.Open(fs.FilePath)
    if err != nil {
        return nil
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
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

func (fs *FileStore) GetStats() RcsStats {
    stats := RcsStats{
        AllCates:    []string{},
        AllTags:     []string{},
        CateTagsMap: map[string][]string{},
        CateNumMap:  map[string]int{},
        TagCatesMap: map[string][]string{},
        TagNumMap:   map[string]int{},
    }

    f, err := os.Open(fs.FilePath)
    if err != nil {
        fmt.Println(err)
        return stats
    }

    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        line := scanner.Text()
        rcs, err := fs.strToCodeSegment(line)
        if err != nil {
            fmt.Println(err)
            continue
        }
        stats.TotalRcsSize ++
        cate := rcs.Category
        tagStr := rcs.Tags
        tagsArr := strings.Split(tagStr, ",")

        cateSize := stats.CateNumMap[cate]
        stats.CateNumMap[cate] = cateSize + 1

        if !ArrContains(stats.AllCates, cate) {
            stats.AllCates = append(stats.AllCates, cate)
        }

        for _, t := range tagsArr {
            if !ArrContains(stats.AllTags, t) {
                stats.AllTags = append(stats.AllTags, t)
            }

            tagsOfCate := stats.CateTagsMap[cate]
            if !ArrContains(tagsOfCate, t) {
                tagsOfCate = append(tagsOfCate, t)
                stats.CateTagsMap[cate] = tagsOfCate
            }

            catesOfTag := stats.TagCatesMap[t]
            if !ArrContains(catesOfTag, cate) {
                catesOfTag = append(catesOfTag, cate)
                stats.TagCatesMap[t] = catesOfTag
            }

            tagSize := stats.TagNumMap[t]
            stats.TagNumMap[t] = tagSize + 1
        }
    }

    return stats
}
