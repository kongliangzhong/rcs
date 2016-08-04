package main

import (
    "fmt"
    "os"
    "strings"
)

const defaultCodeBase = "/opt/rcs-codebase/"
const segFileName = "segfile.rcs"

var segFilePath = defaultCodeBase + segFileName

// keep things simple: category should be one world only. tags can have multiple world, seperated by comma(,).
func main() {
    if len(os.Args) <= 1 {
        printUsage(os.Args)
        os.Exit(-1)
    }

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
        op.ListCates()
    case "list-t":
        op.ListTags()
    case "search":
        codeSeg := parseArgs(os.Args)
        op.Search(codeSeg.Category, codeSeg.Tags)
    case "remove":
        id := os.Args[2]
        fmt.Println("Are you sure to remove code segment with id("+id+")?", "  yes|no")
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
    var content string
    if len(args) > argsLen {
        content = strings.Join(args[argsLen:], " ")
    }

    if args[1] == "search" {
        tagStr = strings.Join(args[argsLen:], ",")
    }

    return CodeSegment{id, cate, tagStr, desc, content}
}

func printUsage(args []string) {
    fmt.Printf("Usage:\n    %s add|update|search|remove|list-c|list-t|merge|append|edit|help\n", args[0])
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
