package main

import (
    //"bufio"
    //"bytes"
    //"encoding/base64"
    //"errors"
    "fmt"
    //"github.com/satori/go.uuid"
    "os"
    "strings"
)

const defaultCodeBase = "/tmp/rcs-codebase/"
const segFileName = "segfile_test.rcs"

var segFilePath = defaultCodeBase + segFileName


// keep things simple: category should be one world only. tags can have multiple world, seperated by comma(,).
func main() {
    // commands: 1. rcs add -t tag1,tag2 -c category -m description content
    // 2. rcs remove id
    // 3. rcs update -i id [-t tag1,tag2 [-c category] [-m desc]] content
    // 4. rcs search [-c category] tag1 tag2
    // 5. rcs list [-c ] [-t] : list all categories or tags
    // TOTO: 6. rcs append -i id content-to-be-appended.
    // TODO: 7. rcs merge id1, id2 ...  // new id created.
    // TODO: 8. rcs edit id : open editor in terminal and edit content of codesegment.
    if len(os.Args) <= 1 {
        printUsage(os.Args)
        os.Exit(-1)
    }

    //fmt.Println("args:", os.Args)

    //id, cate, content, tagStr, desc := parseArgs(os.Args)

    //fmt.Printf("id: %s, cate: %s, tagStr: %s, content: %s", id, cate, tagStr, content)
    op := newOperator(&FileStore{segFilePath})
    switch os.Args[1] {
    case "add":
        codeSeg := parseArgs(os.Args)
        op.Add(codeSeg)
    case "update":
        codeSeg := parseArgs(os.Args)
        op.Update(codeSeg)
    case "append":
        codeSeg := parseArgs(os.Args)
        op.Append(codeSeg.Id, codeSeg.Code)
    case "merge":
        ids := os.Args[2:]
        op.Merge(ids...)
    case "list-c":
        //listc()
    case "list-t":
        //listt()
    case "search":
        codeSeg := parseArgs(os.Args)
        op.Search(codeSeg.Category, codeSeg.Tags)
    case "remove":
        id := os.Args[2]
        fmt.Println("Are you sure to remove code segment with id(" + id + ")?", "  yes|no")
        var response string
        _, err := fmt.Scanln(&response)
        if err != nil {
            fmt.Println(err)
            os.Exit(-1)
        }

        if "YES" == strings.ToUpper(response) {
            op.Remove(id)
        }
    case "edit":
        id := os.Args[2]
        op.Edit(id)
    case "help":
        printUsage(os.Args)
    default:
        printUsage(os.Args)
    }

    if op.err != nil {
        fmt.Println("error:", op.err)
    }
}

func parseArgs(args []string) CodeSegment {
    var ind = func(s string) int {
        for i, a := range args {
            if a == s {
                return i
            }
        }
        return -1
    }

    var argsLen = 2
    var getParam = func(flag string) string {
        if ind_flag := ind(flag); ind_flag > 0 {
            //fmt.Printf("flag:%s, index:%d ", flag, ind_flag)
            if len(args) <= ind_flag+1 {
                fmt.Println("missing parameter value for ", flag)
            }
            argsLen += 2
            return args[ind_flag+1]
        }
        return ""
    }

    id := getParam("-i")
    cate := getParam("-c")
    tagStr := getParam("-t")
    desc := getParam("-m")
    //fmt.Printf("args.len: %d, argsLen: %d", len(args), argsLen)
    var content string
    if len(args) > argsLen {
        content = strings.Join(args[argsLen:], " ")
    }

    return CodeSegment{id, cate, tagStr, desc, content}
}

func printUsage(args []string) {
    fmt.Printf("Usage:\n    %s add|update|list|search|remove|help\n", args[0])
    fmt.Printf("\tadd -t tag1,tag2 -c category -m description content\n")
    fmt.Printf("\tsearch [-c category] tag1 tag2\n")
    fmt.Printf("\tremove id\n")
    fmt.Printf("\tupdate -i id [-t tag1,tag2 [-c category] [-m desc]] content\n")
    fmt.Printf("\tlist-c : list all categories\n")
    fmt.Printf("\tlist-t : list all tags\n")
    fmt.Printf("\tmerge id1 id2 ...\n")
    fmt.Printf("\tappend -i id content\n")
    fmt.Printf("\tedit id")
    fmt.Println()
}

// storage format: id|catetory|t1,t2...|desc_base64|content_base64
// func add(cate string, tagStr string, content string, desc string) error {
//     content = strings.TrimSpace(content)
//     if content == "" {
//         return errors.New("content can not be empty")
//     }

//     if cate == "" && tagStr == "" {
//         return errors.New("category and tagStr can not both empty")
//     }

//     if strings.Contains(cate, "|") || strings.Contains(tagStr, "|") {
//         return errors.New("category and tagStr can not contains '|' charactor")
//     }

//     //descB64 := base64.StdEncoding.EncodeToString([]byte(desc))
//     contentB64 := base64.StdEncoding.EncodeToString([]byte(content))
//     if isDuplicated(segFilePath, []byte(contentB64)) {
//         return errors.New("duplicated content")
//     }

//     id := uuid.NewV4().String()
//     //seg := id + "|" + cate + "|" + tagStr + "|" + descB64 + "|" + contentB64
//     //seg := fldsToLine(id, cate, tagStr, desc, content)
//     return save(id, cate, tagStr, desc, content)
// }

// func fldsToLine(id, cate, tagStr, desc, content string) string {
//     descB64 := base64.StdEncoding.EncodeToString([]byte(desc))
//     contentB64 := base64.StdEncoding.EncodeToString([]byte(content))
//     return id + "|" + cate + "|" + tagStr + "|" + descB64 + "|" + contentB64
// }

// func getById(id string) (seg string, err error) {
//     if id == "" {
//         err = errors.New("id is null")
//         return
//     }

//     f, err := os.Open(segFilePath)
//     if err != nil {
//         return
//     }
//     defer f.Close()

//     scanner := bufio.NewScanner(f)
//     isFound := false
//     for scanner.Scan() {
//         line := scanner.Text()
//         if strings.HasPrefix(line, id) {
//             isFound = true
//             seg = line
//             return
//         }
//     }

//     if !isFound {
//         err = errors.New("can not found code segment with id:" + id)
//         return
//     }

//     return
// }

// func update(id string, cate string, tagStr string, content string, desc string) error {
//     targetSeg, err := getById(id)
//     if err != nil {
//         return err
//     }

//     _id, _cate, _tagStr, _desc, _content := parseSegLine(targetSeg)
//     if id != _id {
//         return errors.New("id not match")
//     }

//     if cate != "" {
//         _cate = cate
//     }

//     if tagStr != "" {
//         _tagStr = tagStr
//     }

//     if desc != "" {
//         _desc = desc
//     }

//     if content != "" {
//         _content = content
//     }

//     remove(id)
//     return save(_id, _cate, _tagStr, _desc, _content)
// }

// func appendContent(id string, extraContent string) error {
//     return nil
// }

// func listc() {

// }

// func listt() {

// }

// func parseSegLine(line string) (id, cate, tagStr, desc, content string) {
//     flds := strings.Split(line, "|")
//     if len(flds) != 5 {
//         fmt.Println("invalid segment format: " + line)
//     }
//     id = flds[0]
//     cate = flds[1]
//     tagStr = flds[2]
//     desc = flds[3]
//     content = flds[4]
//     return
// }

// func search(cate string, tagStr string) {
//     var prtContent = func(line string) {
//         _, _, _, _, contentB64 := parseSegLine(line)
//         bs, err := base64.StdEncoding.DecodeString(contentB64)
//         if err != nil {
//             fmt.Println("error:", err)
//         }
//         content := string(bs)
//         fmt.Println(content)
//     }

//     var prtFull = func(ind int, line string) {
//         fmt.Println(resultDelimiter)
//         id, cate, tagStr, descB64, contentB64 := parseSegLine(line)
//         ctBs, err := base64.StdEncoding.DecodeString(contentB64)
//         if err != nil {
//             fmt.Println("error:", err)
//         }
//         content := string(ctBs)

//         descBs, err := base64.StdEncoding.DecodeString(descB64)
//         if err != nil {
//             fmt.Println("error:", err)
//         }
//         desc := string(descBs)

//         fmt.Println("      id:", id)
//         fmt.Println("category:", cate)
//         fmt.Println("    tags:", tagStr)
//         fmt.Println("    desc:", desc)
//         contentLines := strings.Split(content, "\n")
//         contentLabel := " content:"
//         for i, cline := range contentLines {
//             if i == 0 {
//                 fmt.Println(contentLabel, cline)
//             } else {
//                 fmt.Println(strings.Repeat(" ", len(contentLabel)), cline)
//             }
//         }
//     }

//     tags := strings.Split(tagStr, ",")
//     matches := grepFile(segFilePath, cate, tags)
//     size := len(matches)
//     if size == 0 {
//         fmt.Println("no result found.")
//     } else if size == 1 {
//         prtContent(matches[0])
//     } else {
//         fmt.Printf("found %d matched segments:\n", size)
//         if size > 10 {
//             fmt.Printf("only list 10 result here:\n")
//         }
//         for i, line := range matches {
//             if i == 10 {
//                 break
//             }
//             prtFull(i, line)
//         }
//         fmt.Println(resultDelimiter)
//     }
// }

// TODO: improve this method when necessary.
// func remove(id string) error {
//     f, err := os.Open(segFilePath)
//     if err != nil {
//         return err
//     }
//     defer f.Close()

//     scanner := bufio.NewScanner(f)
//     fLines := []string{}
//     for scanner.Scan() {
//         line := scanner.Text()
//         if strings.HasPrefix(line, id) {
//             continue;
//         }
//         fLines = append(fLines, line)
//     }
//     return replaceFile(fLines)
// }

// func replaceFile(lines []string) error {
//     oldFileName := segFilePath + ".old"
//     os.Remove(oldFileName)
//     err := os.Rename(segFilePath, oldFileName) // do not remove, rename this file instead.
//     if err != nil {
//         return err
//     }

//     f, err := os.OpenFile(segFilePath, os.O_CREATE|os.O_WRONLY, 0660)
//     if err != nil {
//         return err
//     }
//     defer f.Close()
//     for _, line := range lines {
//         f.WriteString(line + "\n")
//     }
//     return nil
// }

// func save(id, cate, tagStr, desc, content string) error {
//     var appendLine = func(line string) error {
//         f, err := os.OpenFile(segFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0660)
//         if err != nil {
//             return err
//         }
//         defer f.Close()

//         _, err = f.WriteString(line + "\n")
//         return err
//     }

//     line := fldsToLine(id, cate, tagStr, desc, content)
//     return appendLine(line)
// }

// func isDuplicated(file string, contentBs []byte) bool {
//     f, err := os.Open(file)
//     if err != nil {
//         fmt.Println(err)
//         return false
//     }
//     defer f.Close()
//     scanner := bufio.NewScanner(f)
//     for scanner.Scan() {
//         if bytes.Contains(scanner.Bytes(), contentBs) {
//             return true
//         }
//     }
//     return false
// }

// func grepFile(file string, cate string, tags []string) []string {
//     var categoryMatch = func(cateInStore string, c string) bool {
//         if c == "" {
//             return true
//         }
//         return strings.HasPrefix(cateInStore, c)
//     }

//     var tagsMatch = func(tagsInStore string, ss []string) bool {
//         if ss == nil || len(ss) == 0 {
//             return true
//         }
//         for _, s := range ss {
//             if !strings.Contains(tagsInStore, strings.ToUpper(s)) { // TODO: improve this logic .
//                 return false
//             }
//         }
//         return true
//     }

//     cate = strings.ToUpper(cate)
//     res := []string{}
//     f, err := os.Open(file)
//     if err != nil {
//         fmt.Println(err)
//     }
//     defer f.Close()
//     scanner := bufio.NewScanner(f)
//     for scanner.Scan() {
//         line := scanner.Text()
//         flds := strings.Split(line, "|")
//         category := flds[1]
//         category = strings.ToUpper(category)
//         tagStr := flds[2]
//         tagStr = strings.ToUpper(tagStr)
//         tagStr = category + "," + tagStr
//         if categoryMatch(category, cate) && tagsMatch(tagStr, tags) {
//             res = append(res, line)
//         }
//     }
//     if err := scanner.Err(); err != nil {
//         fmt.Fprintln(os.Stderr, err)
//     }
//     return res
// }
