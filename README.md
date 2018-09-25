# wtf
command line dictionary written in go

## Installation

```sh
mkdir $HOME/src
cd $HOME/src
git clone https://github.com/master-g/wtf.git
cd wtf
go install
```

**FOR WINDOWS USER**  
change `$HOME` to `%USERPROFILE%`  

## Usage

wtf supports [youdao](http://www.youdao.com/) cloud dictionary only at the moment, other online dictionary might be added in the furture  

1. simple search  

```sh
$ wtf [words...]
```

2. other language (eng, jap)  

```sh
$ wtf -l eng [words...]
```

3. with web translation option  

```sh
$ wtf -w [words...]
```
