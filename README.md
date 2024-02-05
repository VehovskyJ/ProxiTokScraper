# ProxiTokScraper
ProxiTokScraper is a Go program designed for downloading TikTok videos from a user in bulk using the ProxiTok frontend for TikTok. \
You can build the executable yourself from the source code or download it from the releases page

## Building the project:
### Clone the repository:

```bash
git clone https://github.com/VehovskyJ/ProxiTokScraper
cd ProxiTokScraper
```

### Build an executable:

To build an executable for your system use one of the following commands

```bash
make windows
make linux
make arm
make mac
```

## Usage
To download videos from a user just execute the following command and replace the [URL] with the url to of the users profile on ProxiTok

```bash
./ProxiTokScraper-linux-amd64 [URL]
```

The list of public ProxiTok instances can be found [here](https://github.com/pablouser1/ProxiTok/wiki/Public-instances)

Please make sure the instance is fully functional, as some instances may only load the first page.