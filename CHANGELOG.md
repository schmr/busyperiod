# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
### Changed
### Deprecated
### Removed
### Fixed
### Security

## [0.4.0] - 2021-05-12
### Added
- Optimistic variant of EDF-NUVD

## [0.3.3] - 2021-05-07
### Added
- Store counterexample in text file
### Changed
- Allow endless search with `-n 0`
- Check for results concurrently

## [0.3.2] - 2021-05-06
### Changed
- Use workerpool instead of spawning endless goroutines

## [0.3.1] - 2021-05-06
### Added
- Makefile to build `busyp` with revision information
### Changed
- Introduced concurrency in `busyp`
- Use flag package for command line arguments

## [0.2.1] - 2021-05-06
### Added
- [Changelog](./CHANGELOG.md) for humans
- Documentation

## [0.2.0] - 2021-05-06
### Changed
- Generated task sets are implict deadline by default

## [0.1.0] - 2021-05-05
### Added
- First sequential version with limited search
