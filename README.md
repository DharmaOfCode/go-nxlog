# go-nxlog
This is a simple tool that helps security admins in speeding up the installation and configuration of NXLog. At the moment the tool is designed for NXLog configurations for AlienVault USM Anywhere, as it grabs the base configuration file from AlienVault and it sets the SIEM endpoint you choose by passing it to the `-E` argument.

### Usage Examples

**Setup nxlog and set the endpoint to 127.0.0.1 (verbose output):**
```
go-nxlog.exe -E 127.0.0.1 -v
```

**Same as above, specifying the file to download**
```
go-nxlog.exe -E 127.0.0.1 -f nxlog-ce-2.9.1716.msi -v
```
