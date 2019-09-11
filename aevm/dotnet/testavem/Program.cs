using System;
using Google.Protobuf.Collections;
using Getcd;
using aevm;


namespace testavem
{
    class Program
    {
        static void Main(string[] args)
        {
            Console.WriteLine("Hello World!");

            gRPCClient client = new gRPCClient();

            client.connect("47.103.158.190:17000");

            query_global_config_req cat = new query_global_config_req();
            cat.Categories.Add("");

            var queryConfigRes = client.getcd.QueryGlobalConfig(cat);
            Console.WriteLine("queryConfigRes res [" + queryConfigRes.Result + "] entries count[" + queryConfigRes.Entries.Count + "]");

            var protolimit = client.getcd.QueryProtoLimit(new query_proto_limit_req { });
            Console.WriteLine("protolimit res [" + protolimit.Result + "] entries count[" + protolimit.Entries.Count + "]");

            var registryRes = client.getcd.QueryRegistry(new query_registry_req { });
            Console.WriteLine("registryRes res [" + registryRes.Result + "] servers count[" + registryRes.Servers.Count +"] services count["+ registryRes.Services.Count+ "]");
        }
    }
}
