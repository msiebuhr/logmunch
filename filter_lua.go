package logmunch

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shopify/go-lua"
)

func filterLineByLua(prog string, line *LogLine) (bool, error) {
	l := lua.NewState() // new VM instance
	//lua.OpenLibraries(l) // register standard libraries
	//lua.BaseOpen(l)

	// Push timestamp a few times
	l.PushNumber(float64(line.Time.Unix()))
	l.SetGlobal("_time")

	l.PushNumber(float64(line.Time.UnixNano() / 1e6))
	l.SetGlobal("_time_ms")

	// Push name
	l.PushString(line.Name)
	l.SetGlobal("_name")

	// Set key/values from line
	for k, v := range line.Entries {
		k = strings.Replace(k, ".", "_", -1)
		// Can it be passed as a number?
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			l.PushString(v)
		} else {
			l.PushNumber(f)
		}
		l.SetGlobal(k)
	}

	// Make sure this is prefixed with "return", thus creating an anonymous
	// function
	err := lua.DoString(l, "return "+prog)
	if err != nil {
		return false, fmt.Errorf("Cannot parse LUA: `return %s`: %s", prog, err)
	}

	// Call the anonymous function, expecting one return value
	l.ProtectedCall(0, 1, 0)

	// Check that whatever is at top of the array is boolean and then
	// read out it's value

	stackTop := l.Top()
	if typ := l.TypeOf(stackTop); typ != lua.TypeBoolean {
		fmt.Errorf("Expected `%s` to return a boolean, got %s", prog, typ)
	}

	/* DEBUG * /
	for i:=0; i< l.Top()+1; i+=1 {
		fmt.Printf("TypeOf(%d) = %s\n", i, l.TypeOf(i))
		fmt.Printf("ToValue(%d) = %+v\n", i, l.ToValue(i))
		if v, ok := l.ToString(i); ok {
			fmt.Printf("ToString(%d) = `%s`\n", i, v);
		}
		fmt.Printf("IsFunction(%d) = %t\n", i,  l.IsFunction(i))
		fmt.Printf("IsBoolean(%d) = %t\n", i,  l.IsBoolean(i))
		fmt.Printf("ToBoolean(%d) = %t\n", i,  l.ToBoolean(i))
	}

	/*
	if l.ToBoolean(stackTop) == true {
		return line
	} else {
		return nil
	}
	*/
	//fmt.Println(stackTop, l.ToBoolean(stackTop))

	if !l.IsBoolean(stackTop) {
		return false, fmt.Errorf("Program `return %s` returned non-boolean result %s", prog, l.TypeOf(stackTop))
	}
	return l.ToBoolean(stackTop), nil
}

func MakeLuaFilter(prog string) func(*LogLine) *LogLine {
	// TODO(msiebuhr): Make canary run to see if we have valid lua
	return func(line *LogLine) *LogLine {
		keep, err := filterLineByLua(prog, line)
		if err != nil {
			fmt.Println("Error filtering:", err)
			return nil
		}

		if keep {
			return line
		}

		return nil
	}
}
