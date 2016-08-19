package main

import (
    "errors"
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "strings"
    "github.com/satori/go.uuid"
    "strconv"
)

const resultDelimiter = "--------------------------------------------------------"

type Operator struct {
    err   error
    store Store
}

func newOperator(store Store) *Operator {
    return &Operator{nil, store}
}

func (op *Operator) Add(cs CodeSegment) {
    var validateSegment = func() {
        cs.Code = strings.TrimSpace(cs.Code)
        if cs.Code == "" {
            op.err = errors.New("content can not be empty.")
        }

        if cs.Category == "" && cs.Tags == "" {
            op.err = errors.New("category and tags can not be both empty.")
        }

        if strings.Contains(cs.Category, "|") || strings.Contains(cs.Tags, "|") {
            op.err = errors.New("category and tagStr can not contains '|' charactor.")
        }
        return
    }

    if op.err != nil {
        return
    }
    validateSegment()
    if op.err != nil {
        return
    }

    op.err = op.store.Add(cs)
    return
}

func (op *Operator) Update(cs CodeSegment) {
    if cs.Id == "" {
        op.err = errors.New("id is empty")
        return
    }

    op.err = op.store.Update(cs)
}

func (op *Operator) Append(id string, extraContent string) {
    if id == "" || extraContent == "" {
        op.err = errors.New("id or content is nil")
        return
    }

    op.err = op.store.Append(id, extraContent)
}

func (op *Operator) Search(category string, tags string) {
    matchedCs := op.store.Search(category, tags)
    size := len(matchedCs)
    if size > 10 {
        fmt.Println("Found", size, "matched code segments, print first 10 as below:")
    } else {
        fmt.Println("Found", size, "matched code segments, print as below:")
    }
    for i, cs := range matchedCs {
        if i < 10 {
            fmt.Println(resultDelimiter)
            cs.PrintToScreen()
        } else {
            break
        }
    }
    fmt.Println(resultDelimiter)
}

func (op *Operator) Remove(id string) {
    if op.err != nil {
        return
    }
    op.err = op.store.Remove(id)
}

func (op *Operator) Merge(ids ...string) {
    var arrContains = func(arr []string, str string) bool {
        for _, s := range arr {
            if str == s {
                return true
            }
        }
        return false
    }

    var cate string
    var allTags []string
    var desc string
    var code string
    for i, id := range ids {
        cs, err := op.store.GetById(id)
        if err != nil {
            op.err = err
            return
        }

        if i == 0 {
            cate = cs.Category
        }

        desc = desc + "\n" + cs.Desc
        code = code + "\n" + cs.Code

        if cs.Category != cate {
            op.err = errors.New("categorys are not equal, can not merge.")
            return
        }

        tags := strings.Split(cs.Tags, ",")
        for _, t := range tags {
            if !arrContains(allTags, t) {
                allTags = append(allTags, t)
            }
        }
    }

    desc = strings.TrimSpace(desc)
    code = strings.TrimSpace(code)
    allTagsStr := strings.Join(allTags, ",")
    mergedCodeSegment := CodeSegment{"", cate, allTagsStr, desc, code}
    for _, id := range ids {
        op.Remove(id)
    }
    op.Add(mergedCodeSegment)
}

func (op *Operator) Edit(id string) {
    cs, err := op.store.GetById(id)
    if err != nil {
        op.err = err
        return
    }

    tmpDir := os.TempDir()
    tmpFileName := uuid.NewV4().String()
    tmpFile, err := ioutil.TempFile(tmpDir, tmpFileName)
    if err != nil {
        op.err = err
        return
    }
    defer tmpFile.Close()

    cs.PrintToFile(tmpFile.Name())

    path, err := exec.LookPath("vi")
    if err != nil {
        op.err = errors.New("Error while looking for vi: " + err.Error())
        return
    }

    cmd := exec.Command(path, tmpFile.Name())
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    err = cmd.Start()
    if err != nil {
        op.err = err
        return
    }

    err = cmd.Wait()

    if err != nil {
        op.err = err
        return
    }

    //fmt.Println("tmpFile: ", tmpFile.Name())
    err = (&cs).ReadFromFile(tmpFile.Name())
    if err != nil {
        op.err = err
        return
    }
    //cs.PrintToScreen()

    oldId := cs.Id
    cs.Id = ""
    op.Remove(oldId)
    op.Add(cs)
}

func (op *Operator) ListCates() {
    stats := op.store.GetStats()
    head := []string{"INDEX   ", "CATEGORY        ", "RCS-NUM     ", "TAGS"}
    index := 0
    format := fmt.Sprintf("%%-%ds%%-%ds%%-%ds%%-%ds\n", len(head[0]), len(head[1]), len(head[2]), len(head[3]))
    //fmt.Printf("%s%s%s%s\n", head[0], head[1], head[2], head[3])

    lineMax := 50
    for cate, tags := range stats.CateTagsMap {
        index ++
        num := stats.CateNumMap[cate]
        tagLines := []string{}
        line := ""
        for i, tag := range tags {
            if line == "" {
                line = tag
            } else {
                line = line + "," + tag
            }

            if len(line) > lineMax {
                if i != len(tags) - 1 {
                    line = line + ","
                }
                tagLines = append(tagLines, line)
                line = ""
            } else {
                if i == len(tags) - 1 {
                    if strings.HasSuffix(line, ",") {
                        line = line[:len(line)-1]
                    }
                    tagLines = append(tagLines, line)
                }
            }
        }

        for i, tagLine := range tagLines {
            if i == 0 {
                fmt.Printf(format, strconv.Itoa(index), cate, strconv.Itoa(num), tagLine)
            } else {
                formatNewLine := fmt.Sprintf("%%%ds\n", len(head[0]) + len(head[1]) + len(head[2]) + len(tagLine))
                fmt.Printf(formatNewLine, tagLine)
            }
        }

    }
}

func (op *Operator) ListTags() {
    stats := op.store.GetStats()
    head := []string{"INDEX    ", "TAG                    ", "RCS-NUM ", "CATEGORIES    "}
    index := 0
    format := fmt.Sprintf("%%-%ds%%-%ds%%-%ds%%-%ds\n", len(head[0]), len(head[1]), len(head[2]), len(head[3]))
    fmt.Printf("%s%s%s%s\n", head[0], head[1], head[2], head[3])
    for tag, cates := range stats.TagCatesMap {
        index ++
        num := stats.TagNumMap[tag]
        fmt.Printf(format, strconv.Itoa(index), tag, strconv.Itoa(num), strings.Join(cates, ","))
    }
}
