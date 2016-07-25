package main

import (
    "errors"
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "strings"
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
        fmt.Println(resultDelimiter)
        if i < 10 {
            cs.PrintToScreen()
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
            desc = cs.Desc
            code = cs.Code
        } else {
            desc = desc + "\n" + cs.Desc
            code = code + "\n" + cs.Code
        }

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

    allTagsStr := strings.Join(allTags, ",")
    mergedCodeSegment := CodeSegment{"", cate, allTagsStr, desc, code}
    op.Add(mergedCodeSegment)
    for _, id := range ids {
        op.Remove(id)
    }
}

func (op *Operator) Edit(id string) {
    cs, err := op.store.GetById(id)
    if err != nil {
        op.err = err
        return
    }

    tmpDir := os.TempDir()
    tmpFile, err := ioutil.TempFile(tmpDir, cs.Id)
    if err !=nil {
        op.err = err
        return
    }

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

    fmt.Println("tmpFile: ", tmpFile.Name())
    err = (&cs).ReadFromFile(tmpFile.Name())
    if err != nil {
        op.err = err
        return
    }
    cs.PrintToScreen()

    oldId := cs.Id
    cs.Id = ""
    op.Add(cs)
    op.Remove(oldId)
}
