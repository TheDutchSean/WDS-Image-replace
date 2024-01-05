Script created for replacing the WDS Student image automaticly with the use of a windows task.

configure the script using config.json

    "server":"SETS THE NAME OF THE WDS SERVER",
    "image":{
        "group":"SETS THE NAME OF THE IMAGE GROUP",
        "name":"SETS THE NAME OF THE IMAGE",
        "path":"SETS PATH TO THE NEW IMAGE.wim"
    },
    "log":{
        "name":"log.txt",
        "path":""
    }

script written in GoLang