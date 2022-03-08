//go:generate msgp -tests=false
//msgp:ignore Repository
package xdb

// Deprecated
// 请使用 xdb.Model
type Entity = Model

// Deprecated
// 请使用 xdb.Dao
type Repository = Dao
