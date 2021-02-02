package schema

import (
	"crypto/sha1"
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/jinzhu/inflection"
)

// Namer namer interface
type Namer interface {
	TableName(table string) string
	ColumnName(table, column string) string
	JoinTableName(joinTable string) string
	RelationshipFKName(Relationship) string
	CheckerName(table, column string) string
	IndexName(table, column string) string
}

// Replacer replacer interface like strings.Replacer
type Replacer interface {
	Replace(name string) string
}

// NamingStrategy tables, columns naming strategy
type NamingStrategy struct {
	TablePrefix   string
	SingularTable bool
	NameReplacer  Replacer
}

// TableName convert string to table name
func (ns NamingStrategy) TableName(str string) string {
	if ns.SingularTable {
		return ns.TablePrefix + ns.toDBName(str)
	}
	return ns.TablePrefix + inflection.Plural(ns.toDBName(str))
}

// ColumnName convert string to column name
func (ns NamingStrategy) ColumnName(table, column string) string {
	return ns.toDBName(column)
}

// JoinTableName convert string to join table name
func (ns NamingStrategy) JoinTableName(str string) string {
	if strings.ToLower(str) == str {
		return ns.TablePrefix + str
	}

	if ns.SingularTable {
		return ns.TablePrefix + ns.toDBName(str)
	}
	return ns.TablePrefix + inflection.Plural(ns.toDBName(str))
}

// RelationshipFKName generate fk name for relation
func (ns NamingStrategy) RelationshipFKName(rel Relationship) string {
	return ns.formatName("fk", rel.Schema.Table, ns.toDBName(rel.Name))
}

// CheckerName generate checker name
func (ns NamingStrategy) CheckerName(table, column string) string {
	return ns.formatName("chk", table, column)
}

// IndexName generate index name
func (ns NamingStrategy) IndexName(table, column string) string {
	return ns.formatName("idx", table, ns.toDBName(column))
}

func (ns NamingStrategy) formatName(prefix, table, name string) string {
	formatedName := strings.Replace(fmt.Sprintf("%v_%v_%v", prefix, table, name), ".", "_", -1)

	if utf8.RuneCountInString(formatedName) > 64 {
		h := sha1.New()
		h.Write([]byte(formatedName))
		bs := h.Sum(nil)

		formatedName = fmt.Sprintf("%v%v%v", prefix, table, name)[0:56] + string(bs)[:8]
	}
	return formatedName
}

var (
	smap sync.Map
	// https://github.com/golang/lint/blob/master/lint.go#L770
	commonInitialisms         = []string{"API", "ASCII", "CPU", "CSS", "DNS", "EOF", "GUID", "HTML", "HTTP", "HTTPS", "ID", "IP", "JSON", "LHS", "QPS", "RAM", "RHS", "RPC", "SLA", "SMTP", "SSH", "TLS", "TTL", "UID", "UI", "UUID", "URI", "URL", "UTF8", "VM", "XML", "XSRF", "XSS"}
	commonInitialismsReplacer *strings.Replacer
)

func init() {
	var commonInitialismsForReplacer []string
	for _, initialism := range commonInitialisms {
		commonInitialismsForReplacer = append(commonInitialismsForReplacer, initialism, strings.Title(strings.ToLower(initialism)))
	}
	commonInitialismsReplacer = strings.NewReplacer(commonInitialismsForReplacer...)
}

// reset should be called before each unit test. It clears out the cached names from smap. TODO: make smap part of NamingStrategy instead of a global singleton.
func reset() {
	smap = sync.Map{}
}

func (ns NamingStrategy) toDBName(name string) string {
	if name == "" {
		return ""
	} else if v, ok := smap.Load(name); ok {
		return v.(string)
	}

	if ns.NameReplacer != nil {
		name = ns.NameReplacer.Replace(name)
	}

	var (
		value                          = commonInitialismsReplacer.Replace(name)
		buf                            strings.Builder
		lastCase, nextCase, nextNumber bool // upper case == true
		curCase                        = value[0] <= 'Z' && value[0] >= 'A'
	)

	for i, v := range value[:len(value)-1] {
		nextCase = value[i+1] <= 'Z' && value[i+1] >= 'A'
		nextNumber = value[i+1] >= '0' && value[i+1] <= '9'

		if curCase {
			if lastCase && (nextCase || nextNumber) {
				buf.WriteRune(v + 32)
			} else {
				if i > 0 && value[i-1] != '_' && value[i+1] != '_' {
					buf.WriteByte('_')
				}
				buf.WriteRune(v + 32)
			}
		} else {
			buf.WriteRune(v)
		}

		lastCase = curCase
		curCase = nextCase
	}

	if curCase {
		if !lastCase && len(value) > 1 {
			buf.WriteByte('_')
		}
		buf.WriteByte(value[len(value)-1] + 32)
	} else {
		buf.WriteByte(value[len(value)-1])
	}
	ret := buf.String()
	smap.Store(name, ret)
	return ret
}
