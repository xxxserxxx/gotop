# Current state

First, there were go plugins. This turned out to be impractical due to the limitations in plugins making them unsuitable for use outside of a small, strict, and (one could argue) useless use case.

Then I tried external static extensions. This approach used a trick to copy and modify the gotop main executable, which then imported it's own packages from upstream.  This worked, but was awkward and required several steps to build.

Currently, as I've only written two modules since I started down this path, and there's no clean, practical solution yet in Go, I've folded the extensions into the main codebase. This means there's no programmatic extension mechanism for gotop.
