package common

/*

Provides a basic mechanism to help components do data migration easily

reader
writer
domain
version

Schemas are stored by "domain", which allows to segregate data migration between sub-components

Vx-1 : fully transtionned to Vx-1 version, writes AND reads are done with the Vx-1 schema
Tx   : transitionning to Vx version, writes are done with Vx schema, but reads are an aggregation of Vx and Vx-1 data. Meanwhile the migration process occurs...
Vx   : fully transtionned to Vx version, writes AND reads are done with the Vx schema



Users provides those methods which are called by MyOwnCluster :

GetModernVersion(): number : the user code should return the best fitted most recent schema version. My Own Cluster will undertake a data migration if there is need for it
MigrateData(x: target version):Status : the user code should pick one data in format x-1 and migrate it to format x. If there is no data to migrate, the user code should return NULL to signify that data migration to version 'x' is done for the domain.




MyOwnCluster provides this API:

RegisterDomainHandler(name: String) : registers a code for handling data migration of a domain
GetDomainWriteVersion():String : returns the schema version the user code should use when storing a data. This is this version that is implicitely used as a reference in the IsDomainBackwardCompatible()
IsDomainBackwardCompatible():boolean : returns true if the data is actual stored in both the domain's WRITE_VERSION and its predecessor, or if all data is at the domain's WRITE_VERSION
Hold(domain:String):Unholder : all read and writes to a domain should be contained between a Hold and Unhold() call. Acts like a critical section, so only one thread at a time. This is important to maintain database migrations' consistency.
Unholder.Unhold():void : releases a domain



A user code should typically write data in the format indicated by `GetDomainWriteVersion()`.
For reads, the the user code should read data in the format indicated by `GetDomainWriteVersion()` and call `IsDomainBackwardCompatible()` to know if it should also merge its results with data from the x-1 version of the schema.


*/
