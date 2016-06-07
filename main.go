package main

import (
    "bufio"
    "encoding/base64"
    "errors"
    "fmt"
    "github.com/satori/go.uuid"
    "io/ioutil"
    "log"
    "os"
    "strings"
)

const defaultCodeBase = "/opt/rcs_codebase/"
const segFileName = "segfile.rcs"

var segFilePath = defaultCodeBase + segFileName

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

    id, cate, content, tags := parseArgs(os.Args)
    switch os.Args[1] {
    case "add":
        log.Fatal(add(cate, tags, content))
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

func parseArgs(args []string) (id string, cate string, content string, tags string) {
    var ind = func(s string) int {
        for i, a := range args {
            if a == s {
                return i
            }
        }
        return -1
    }

    if ind_i := ind("-i"); ind_i > 0 {
        if len(args) <= ind_i+1 {
            log.Fatal("missing parameter value for -i")
        }
        id = args[ind_i+1]
    }

    return
}

func printUsage(args []string) {
    fmt.Printf("Usage: %s add|update|list|search|remove|help", args[0])
}

// storage format: id|catetory|t1,t2...|content_base64
func add(cate string, tags string, content string) error {
    content = strings.TrimSpace(content)
    if content == "" {
        return errors.New("content can not be empty")
    }

    if cate == "" && tags == "" {
        return errors.New("category and tags can not be both empty")
    }

    content_b64 := base64.StdEncoding.EncodeToString([]byte(content))
    id := uuid.NewV4().String()
    seg := id + "|" + cate + "|" + tags + "|" + content_b64
    return save(seg)
}

func update(id string, cate string, tags string, content string) error {
    return nil
}

func list(cate string, tagStr string) {

}

func search(cate string, tagStr string) {
    tags := strings.Split(tagStr, ",")
    grepFile(segFilePath, cate, tags)
}

func remove(id string) error {
    return nil
}

func save(seg string) error {
    f := defaultCodeBase + segFileName
    return ioutil.WriteFile(f, []byte(seg), 0660)
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
        log.Fatal(err)
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
