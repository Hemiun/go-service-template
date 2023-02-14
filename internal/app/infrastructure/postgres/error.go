package postgres

import "errors"

var (
	// ErrTxNotFound - "can't get tx from context: nil value got" error
	ErrTxNotFound = errors.New("can't get tx from context: nil value got")

	// ErrTxTypeConversation - "can't get tx from context: conversion error"
	ErrTxTypeConversation = errors.New("can't get tx from context: conversion error")
)
