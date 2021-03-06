[![Build Status](https://travis-ci.org/lca1/unlynx.svg?branch=master)](https://travis-ci.org/LCA1/UnLynx) [![Go Report Card](https://goreportcard.com/badge/github.com/lca1/unlynx)](https://goreportcard.com/report/github.com/lca1/unlynx) [![Coverage Status](https://coveralls.io/repos/github/lca1/unlynx/badge.svg?branch=master)](https://coveralls.io/github/lca1/unlynx?branch=master)
# UnLynx 
UnLynx is a library for simulating a privacy-preserving data sharing tool. It offers a series of independent protocols that when combined offer a verifiably-secure and safe way to share sensitive data (e.g., medical data).  

UnLynx is developed by lca1 (Laboratory for Communications and Applications in EPFL) in collaboration with DeDis (Laboratory for Decentralized and Distributed Systems).  

## Documentation

* The UnLynx library does an intensive use of [Overlay-network (ONet) library](https://github.com/dedis/onet) and of the [Advanced Crypto (kyber) library](https://github.com/dedis/kyber).
* For more information regarding the underlying architecture please refer to the stable version of ONet `github.bom/dedis/onet`
* To check the code organisation, have a look at [Layout](https://github.com/lca1/unlynx/wiki/Layout)
* For more information on how to run our protocols, services, simulations and apps, go to [Running UnLynx](https://github.com/lca1/unlynx/wiki/Running-UnLynx)

## Getting Started

To use the code of this repository you need to:

- Install [Golang](https://golang.org/doc/install)
- [Recommended] Install [IntelliJ IDEA](https://www.jetbrains.com/idea/) and the GO plugin
- Set [`$GOPATH`](https://golang.org/doc/code.html#GOPATH) to point to your workspace directory
- Add `$GOPATH/bin` to `$PATH`
- Git clone this repository to $GOPATH/src `git clone https://github.com/lca1/unlynx.git` or...
- go get repository: `go get github.com/lca1/unlynx`

## Version

We have a development and a stable version. The `master`-branch in `github.com/lca1/unlynx` is the development version that works but can have incompatible changes.

Use one of the latest tags `v1.2b-alpha` that are stable and have no incompatible changes.

**Very Important!!** 

Due to the current changes being made to [onet](https://github.com/dedis/onet) and [kyber](https://github.com/dedis/kyber) (release of v3) you must revert back to previous commits for these two libraries if you want UnLynx to work. This will change in the near future. 

```bash
cd $GOPATH/src/dedis/onet/
git checkout 5796104343ef247e2eed58e573f68c566db2136f

cd $GOPATH/src/dedis/kyber/
git checkout f55fec5463cda138dfc7ff15e4091d12c4ddcbfe
```

## License

UnLynx is licensed under a End User Software License Agreement ('EULA') for non-commercial use. If you need more information, please contact us.

## Contact
You can contact any of the developers for more information or any other member of [lca1](http://lca.epfl.ch/people/lca1/):

* [David Froelicher](https://github.com/froelich) (PHD student) - david.froelicher@epfl.ch
* [Patricia Egger](https://github.com/pegger) (Security Consultant at Deloitte) - paegger@deloitte.ch
* [Joao Andre Sa](https://github.com/JoaoAndreSa) (Software Engineer) - joao.gomesdesaesousa@epfl.ch
* [Christian Mouchet](https://github.com/ChristianMct) (PHD student) - christian.mouchet@epfl.ch
