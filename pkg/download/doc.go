// Package download implements package download functionality for aqua.
// It provides an HTTP client for downloading packages with support for
// progress tracking, retry logic, and cache management. This package handles
// the reliable retrieval of CLI tools from various sources while managing
// network failures and optimizing performance through caching.
package download