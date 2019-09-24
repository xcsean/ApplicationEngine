using System;
using System.Threading.Tasks;
using Grpc.Core;
using Getcd;

namespace aevm
{
    public class GetcdServiceImpl : GetcdService.GetcdServiceBase
    {
        public override Task<query_registry_rsp> QueryRegistry(query_registry_req request, ServerCallContext context)
        {
            // for Debug ...
            Console.WriteLine("on QueryRegistry");

            query_registry_rsp rsp = new query_registry_rsp();
            rsp.Result = 0;
            return Task.FromResult(rsp);
        }
        public override Task<query_global_config_rsp> QueryGlobalConfig(query_global_config_req request, ServerCallContext context)
        {
            // for Debug ...
            Console.WriteLine("on QueryGlobalConfig");

            query_global_config_rsp rsp = new query_global_config_rsp();
            rsp.Result = 0;
            return Task.FromResult(rsp);
        }
        public override Task<query_proto_limit_rsp> QueryProtoLimit(query_proto_limit_req request, ServerCallContext context)
        {
            // for Debug ...
            Console.WriteLine("on QueryProtoLimit");

            query_proto_limit_rsp rsp = new query_proto_limit_rsp();
            rsp.Result = 0;
            return Task.FromResult(rsp);
        }
    }

    public class gRPCServer
    {
        protected int       _port = 9007;
        protected string    _host = "localhost";

        protected GetcdServiceImpl  _getService = null;
        protected Server            _grpcServer = null;

        public gRPCServer()
        {

        }

        public int port
        {
            get
            {
                return _port;
            }
        }

        public string host
        {
            get
            {
                return _host;
            }
        }

        public void run(string host, int port)
        {
            _port = port;
            _host = host;

            _getService = new GetcdServiceImpl();

            _grpcServer = new Server
            {
                Services = { GetcdService.BindService(_getService) },
                Ports = { new ServerPort(_host, _port, ServerCredentials.Insecure) }
            };
            _grpcServer.Start();
        }

        public void stop()
        {
            _grpcServer.ShutdownAsync().Wait();
        }
    }
}
