* please refer to config.json: 

>* Strategy : When the field here is' local', it means to run locally. If you run on the server, you need to change this.
>* RemoteStrategy : address of the cloud bootnode
>* PublicAdd : public address for cloud
>* PrivateAdd : private address for cloud
>* addresses : addresses for the users
>* transIndex : this is the amount of transactions that will be stored in one Block
>* maximumDiffculty: this is the maximum Difficulty that the Block will have. The first Block(besides the genesis Block)
>will have difficulty of 1, and will be increasing until the amount of the maximum achieved.

* to run the file please do:

~~~shell
go build
./POW_ghostProtocol 
~~~

