package main

import (
    "bufio"
    "bytes"
    "encoding/base64"
    "errors"
    "fmt"
    "github.com/satori/go.uuid"
    //"log"
    "os"
    "strings"
)

const defaultCodeBase = "/opt/rcs-codebase/"
const segFileName = "segfile.rcs"
const resultDelimiter = "--------------------------------------------------------"

var segFilePath = defaultCodeBase + segFileName

// keep things simple: category should be one world only. tags can have multiple world, concat by comma(,).
func main() {
    // commands: 1. rcs add -t tag1,tag2 -c category content
    // 2. rcs remove id
    // 3. rcs update -i id [-t tag1,tag2 [-c category]] content
    // 4. rcs search [-c category] tag1 tag2
    // 5. rcs list [-c category [-t t1,t2]]
    if len(os.Args) <= 1 {
        printUsage(os.Args)
        os.Exit(-1)
    }

    //fmt.Println("args:", os.Args)

    id, cate, content, tags := parseArgs(os.Args)
    //fmt.Printf("id: %s, cate: %s, tags: %s, content: %s", id, cate, tags, content)
    switch os.Args[1] {
    case "add":
        err := add(cate, tags, content)
        if err != nil {
            fmt.Println("error:", err)
        }
    case "update":
        update(id, cate, tags, content)
    case "list":
        list(cate, tags)
    case "search":
        search(cate, tags)
    case "remove":
        remove(id)
    case "help":
        printUsage(os.Args)
    default:
        printUsage(os.Args)
    }
}

func parseArgs(args []string) (id string, cate string, content string, tagStr string) {
    var ind = func(s string) int {
        for i, a := range args {
            if a == s {
                return i
            }
        }
        return -1
    }

    var hasId, hasCate, hasTags bool
    var argsLen = 2
    var getParam = func(flag string) string {
        if ind_flag := ind(flag); ind_flag > 0 {
            //fmt.Printf("flag:%s, index:%d ", flag, ind_flag)
            if len(args) <= ind_flag+1 {
                fmt.Println("missing parameter value for ", flag)
            }
            switch flag {
            case "-i":
                hasId = true
                argsLen += 2
            case "-c":
                hasCate = true
                argsLen += 2
            case "-t":
                hasTags = true
                argsLen += 2
            }
            return args[ind_flag+1]
        }
        return ""
    }

    id = strings.ToLower(getParam("-i"))
    cate = strings.ToLower(getParam("-c"))
    tagStr = strings.ToLower(getParam("-t"))

    //fmt.Printf("args.len: %d, argsLen: %d", len(args), argsLen)
    if len(args) > argsLen {
        content = args[argsLen]
    }

    return
}

func printUsage(args []string) {
    fmt.Printf("Usage: %s add|update|list|search|remove|help", args[0])
}

// storage format: id|catetory|t1,t2...|content_base64
func add(cate string, tagStr string, content string) error {
    content = strings.TrimSpace(content)
    if content == "" {
        return errors.New("content can not be empty")
    }

    if cate == "" && tagStr == "" {
        return errors.New("category and tagStr can not both empty")
    }

    if strings.Contains(cate, "|") || strings.Contains(tagStr, "|") {
        return errors.New("category and tagStr can not contains '|' charactor")
    }

    content_b64 := base64.StdEncoding.EncodeToString([]byte(content))
    if isDuplicated(segFilePath, []byte(content_b64)) {
        return errors.New("duplicated content")
    }

    id := uuid.NewV4().String()
    seg := id + "|" + cate + "|" + tagStr + "|" + content_b64
    return save(seg)
}

func update(id string, cate string, tagStr string, content string) error {
    return nil
}

func list(cate string, tagStr string) {
    // tags := strings.Split(tagStr, ",")
    // matches := grepFile(segFilePath, cate, tags)
}

func search(cate string, tagStr string) {
    var prtContent = func(line string) {
        flds := strings.Split(line, "|")
        contentB64 := flds[3]
        bs, err := base64.StdEncoding.DecodeString(contentB64)
        if err != nil {
            fmt.Println("error:", err)
        }
        fmt.Println(string(bs))
    }

    var prtFull = func(ind int, line string) {
        fmt.Println(resultDelimiter)
        flds := strings.Split(line, "|")
        id := flds[0]
        cate := flds[1]
        tagStr := flds[2]
        contentB64 := flds[3]
        bs, err := base64.StdEncoding.DecodeString(contentB64)
        if err != nil {
            fmt.Println("error:", err)
        }
        content := string(bs)
        fmt.Println("      id:", id)
        fmt.Println("category:", cate)
        fmt.Println("    tags:", tagStr)
        fmt.Println(" content:", content)
    }

    tags := strings.Split(tagStr, ",")
    matches := grepFile(segFilePath, cate, tags)
    size := len(matches)
    if size == 0 {
        fmt.Println("no result found.")
    } else if size == 1 {
        prtContent(matches[0])
    } else {
        fmt.Printf("found %d matched segments:\n", size)
        if size > 10 {
            fmt.Printf("only list 10 result here:\n")
        }
        for i, line := range matches {
            if i == 10 {
                break
            }
            prtFull(i, line)
        }
        fmt.Println(resultDelimiter)
    }
}

func remove(id string) error {
    return nil
}

func save(seg string) error {
    f, err := os.OpenFile(segFilePath, os.O_APPEND|os.O_WRONLY, 0660)
    if err != nil {
        return err
    }

    defer f.Close()

    _, err = f.WriteString(seg + "\n")
    return err

    // f := defaultCodeBase + segFileName
    // return ioutil.WriteFile(f, []byte(seg), 0660)
}

func isDuplicated(file string, contentBs []byte) bool {
    f, err := os.Open(file)
    if err != nil {
        fmt.Println(err)
        return true
    }
    defer f.Close()
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        if bytes.Contains(scanner.Bytes(), contentBs) {
            return true
        }
    }
    return false
}

func grepFile(file string, cate string, tags []string) []string {
    var categoryMatch = func(src string, c string) bool {
        if c == "" {
            return true
        }
        return strings.Contains(src, c)
    }

    var tagsMatch = func(src string, ss []string) bool {
        if ss == nil || len(ss) == 0 {
            return true
        }
        for _, s := range ss {
            if !strings.Contains(src, s) {
                return false
            }
        }
        return true
    }

    res := []string{}
    f, err := os.Open(file)
    if err != nil {
        fmt.Println(err)
    }
    defer f.Close()
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        line := scanner.Text()
        flds := strings.Split(line, "|")
        category := flds[1]
        tagStr := flds[2]
        if categoryMatch(category, cate) && tagsMatch(tagStr, tags) {
            res = append(res, line)
        }
    }
    if err := scanner.Err(); err != nil {
        fmt.Fprintln(os.Stderr, err)
    }
    return res
}
