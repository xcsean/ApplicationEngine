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
            cat.Categories.Add("global");

            long start = DateTime.UtcNow.Ticks;

            var queryConfigRes = client.getcd.QueryGlobalConfig(cat);

            long qctN = DateTime.UtcNow.Ticks;
            long qct = qctN - start;

            var protolimit = client.getcd.QueryProtoLimit(new query_proto_limit_req { });

            long pltN = DateTime.UtcNow.Ticks;
            long plt = pltN - qctN;

            var registryRes = client.getcd.QueryRegistry(new query_registry_req { });

            long rrtN = DateTime.UtcNow.Ticks;
            long rrt = rrtN - pltN;

            Console.WriteLine("queryConfigRes tm:"+ qct + " res [" + queryConfigRes.Result + "] entries count[" + queryConfigRes.Entries.Count + "] entries:"+ queryConfigRes.Entries.ToString());

            Console.WriteLine("protolimit tm:"+ plt + "res [" + protolimit.Result + "] entries count[" + protolimit.Entries.Count + "]");

            Console.WriteLine("registryRes tm:"+ rrt + " res [" + registryRes.Result + "] servers count[" + registryRes.Servers.Count +"] services count["+ registryRes.Services.Count+ "]");
        }
    }
}
