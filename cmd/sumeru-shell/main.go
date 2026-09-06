// Command sumeru-shell is an interactive ORM REPL for developers.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"sumeru/core/applog"
	"sumeru/core/orm"
	"sumeru/core/server/cliboot"
)

func main() {
	cliboot.StripLeadingArgsSeparator()
	configPath := flag.String("c", "sumeru.conf", "Path to config file (INI)")
	uidFlag := flag.Int("uid", 1, "Security uid for ORM calls")
	flag.Parse()

	ctx, err := cliboot.Init(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init: %v\n", err)
		os.Exit(1)
	}
	ctx = cliboot.ContextWithUID(ctx, *uidFlag)

	fmt.Println("Sumeru shell — commands: search, read, create, write, domain, models, quit")
	sc := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("sumeru> ")
		if !sc.Scan() {
			break
		}
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		parts := splitFields(line)
		cmd := strings.ToLower(parts[0])
		switch cmd {
		case "quit", "exit", "q":
			return
		case "help", "?":
			printHelp()
		case "models":
			for name := range orm.Registry {
				fmt.Println(name)
			}
		case "search":
			if len(parts) < 2 {
				fmt.Println("usage: search MODEL [limit]")
				continue
			}
			limit := 10
			if len(parts) >= 3 {
				parsed, err := strconv.Atoi(parts[2])
				if err != nil || parsed <= 0 {
					fmt.Fprintln(os.Stderr, "error: invalid limit")
					continue
				}
				limit = parsed
			}
			rows, err := orm.SearchLimit(ctx, parts[1], nil, limit)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				continue
			}
			printJSON(rows)
		case "read":
			if len(parts) < 3 {
				fmt.Println("usage: read MODEL ID")
				continue
			}
			id, err := strconv.Atoi(parts[2])
			if err != nil || id <= 0 {
				fmt.Fprintln(os.Stderr, "error: invalid id")
				continue
			}
			row, err := orm.SearchOne(ctx, parts[1], map[string]interface{}{"id": id})
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				continue
			}
			printJSON(row)
		case "create":
			if len(parts) < 3 {
				fmt.Println(`usage: create MODEL {"field":"value"}`)
				continue
			}
			var vals map[string]interface{}
			if err := json.Unmarshal([]byte(strings.Join(parts[2:], " ")), &vals); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				continue
			}
			inst, ok := orm.Registry[parts[1]]
			if !ok || inst == nil {
				fmt.Fprintln(os.Stderr, "error: unknown model")
				continue
			}
			id, err := orm.Create(ctx, inst, vals)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				continue
			}
			fmt.Println("id:", id)
		case "write":
			if len(parts) < 4 {
				fmt.Println(`usage: write MODEL ID {"field":"value"}`)
				continue
			}
			id, err := strconv.Atoi(parts[2])
			if err != nil || id <= 0 {
				fmt.Fprintln(os.Stderr, "error: invalid id")
				continue
			}
			var vals map[string]interface{}
			if err := json.Unmarshal([]byte(strings.Join(parts[3:], " ")), &vals); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				continue
			}
			if err := orm.UpdateRecordByID(ctx, parts[1], id, vals); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				continue
			}
			fmt.Println("ok")
		case "domain":
			if len(parts) < 3 {
				fmt.Println(`usage: domain MODEL [[ "field","=","x" ]]`)
				continue
			}
			raw := strings.Join(parts[2:], " ")
			dom, err := orm.ParseDomainJSON(raw)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				continue
			}
			where, args, err := orm.BuildWhereWithRecordRules(ctx, orm.SecurityUID(ctx), parts[1], "read", dom)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				continue
			}
			fmt.Println("WHERE:", where)
			printJSON(map[string]interface{}{"args": args})
		default:
			fmt.Println("unknown command; type help")
		}
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
	}
}

func printHelp() {
	fmt.Println(`Commands:
  models                          List registered models
  search MODEL [limit]              Search records
  read MODEL ID                   Read one record
  create MODEL {json}             Create record
  write MODEL ID {json}           Update record
  domain MODEL [domain-json]      Preview SQL WHERE for domain
  quit                            Exit`)
}

func splitFields(s string) []string {
	var out []string
	var cur strings.Builder
	inQuote := false
	for _, r := range s {
		switch {
		case r == '"':
			inQuote = !inQuote
			cur.WriteRune(r)
		case r == ' ' && !inQuote:
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

func printJSON(v interface{}) {
	b, err := json.MarshalIndent(applog.ScrubValue("", v), "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return
	}
	fmt.Println(string(b))
}
