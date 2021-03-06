// Copyright (C) MongoDB, Inc. 2017-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package command

import (
	"context"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/core/description"
	"github.com/mongodb/mongo-go-driver/core/options"
	"github.com/mongodb/mongo-go-driver/core/result"
	"github.com/mongodb/mongo-go-driver/core/wiremessage"
)

// FindOneAndDelete represents the findOneAndDelete operation.
//
// The findOneAndDelete command deletes a single document that matches a query and returns it.
type FindOneAndDelete struct {
	NS    Namespace
	Query *bson.Document
	Opts  []options.FindOneAndDeleteOptioner

	result result.FindAndModify
	err    error
}

// Encode will encode this command into a wire message for the given server description.
func (f *FindOneAndDelete) Encode(desc description.SelectedServer) (wiremessage.WireMessage, error) {
	if err := f.NS.Validate(); err != nil {
		return nil, err
	}

	command := bson.NewDocument(
		bson.EC.String("findAndModify", f.NS.Collection),
		bson.EC.SubDocument("query", f.Query),
		bson.EC.Boolean("remove", true),
	)

	for _, option := range f.Opts {
		if option == nil {
			continue
		}
		err := option.Option(command)
		if err != nil {
			return nil, err
		}
	}

	return (&Command{DB: f.NS.DB, Command: command, isWrite: true}).Encode(desc)
}

// Decode will decode the wire message using the provided server description. Errors during decoding
// are deferred until either the Result or Err methods are called.
func (f *FindOneAndDelete) Decode(desc description.SelectedServer, wm wiremessage.WireMessage) *FindOneAndDelete {
	rdr, err := (&Command{}).Decode(desc, wm).Result()
	if err != nil {
		f.err = err
		return f
	}

	f.result, f.err = unmarshalFindAndModifyResult(rdr)
	return f
}

// Result returns the result of a decoded wire message and server description.
func (f *FindOneAndDelete) Result() (result.FindAndModify, error) {
	if f.err != nil {
		return result.FindAndModify{}, f.err
	}
	return f.result, nil
}

// Err returns the error set on this command.
func (f *FindOneAndDelete) Err() error { return f.err }

// RoundTrip handles the execution of this command using the provided wiremessage.ReadWriter.
func (f *FindOneAndDelete) RoundTrip(ctx context.Context, desc description.SelectedServer, rw wiremessage.ReadWriter) (result.FindAndModify, error) {
	wm, err := f.Encode(desc)
	if err != nil {
		return result.FindAndModify{}, err
	}

	err = rw.WriteWireMessage(ctx, wm)
	if err != nil {
		return result.FindAndModify{}, err
	}
	wm, err = rw.ReadWireMessage(ctx)
	if err != nil {
		return result.FindAndModify{}, err
	}
	return f.Decode(desc, wm).Result()
}
