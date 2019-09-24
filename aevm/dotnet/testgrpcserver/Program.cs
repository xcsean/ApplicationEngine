using System;
using Google.Protobuf.Collections;
using Getcd;
using aevm;

namespace testgrpcserver
{
    class Program
    {
        static void Main(string[] args)
        {
            Console.WriteLine("start grpc server!");

            gRPCServer server = new gRPCServer();

            server.run("localhost", 9007);

            Console.WriteLine("gRPC server listening on port " + server.port);
            Console.WriteLine("press any key to quit...");
            Console.ReadKey();

            server.stop();
        }
    }
}
