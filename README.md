# reimport
Golang Import Patching tool

## description
reimport is a quick & dirty tool for generating a patch(1) for import lines you want to change
in large Go projects

it uses go/parser in order to accurately find the import lines (hence not patching code lines with matching string).

## build / install

    go get github.com/unix4fun/reimport

## example usage
    
    reimport -m hash -r myhash /path/to/dir > import.patch

## example output:


    $ reimport -m crypto -r tagada ~/dev/ic/ 
    --- /home/eau/dev/ic/iccp/pack_ac.go
    +++ /home/eau/dev/ic/iccp/pack_ac.go
    @@ -10,1 +10,1 @@
    -	"golang.org/x/crypto/nacl/secretbox"
    +	"golang.org/x/tagada/nacl/secretbox"
    --- /home/eau/dev/ic/iccp/pack_kx.go
    +++ /home/eau/dev/ic/iccp/pack_kx.go
    @@ -11,1 +11,1 @@
    -	"golang.org/x/crypto/nacl/box" // nacl is now here.
    +	"golang.org/x/tagada/nacl/box" // nacl is now here.
    --- /home/eau/dev/ic/iccp/protocol.go
    +++ /home/eau/dev/ic/iccp/protocol.go
    @@ -9,1 +9,1 @@
    -	"golang.org/x/crypto/nacl/secretbox" // nacl is now here.
    +	"golang.org/x/tagada/nacl/secretbox" // nacl is now here.
    --- /home/eau/dev/ic/icjs/msgct.go
    +++ /home/eau/dev/ic/icjs/msgct.go
    @@ -6,1 +6,1 @@
    -	"crypto/rand"
    +	"tagada/rand"
    ... (stripped)...


## bleh
enjoy!
