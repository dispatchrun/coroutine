[![build](https://github.com/stealthrocket/coroutine/actions/workflows/build.yml/badge.svg)](https://github.com/stealthrocket/coroutine/actions/workflows/build.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/stealthrocket/coroutine.svg)](https://pkg.go.dev/github.com/stealthrocket/coroutine)
[![Apache 2 License](https://img.shields.io/badge/license-Apache%202-blue.svg)](LICENSE)

# Durable Coroutines

Coroutines are functions that can be suspended and resumed. _Durable_ coroutines
are functions that can be suspended, serialized and resumed in another process.

This repository contains a durable coroutine compiler and runtime library for Go.
