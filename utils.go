package main

func ArrContains(strArr []string, s string) bool {
    for _, str := range strArr {
        if s == str {
            return true
        }
    }

    return false
}
