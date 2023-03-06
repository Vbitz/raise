using Go = import "/go.capnp";
@0xc1dd04a4b598e212;

$Go.package("proto");
$Go.import("github.com/Vbitz/raise/v2/pkg/proto");

interface Service {
    ping @0 (name :Text);
}

interface ClientService extends(Service) {
}

interface WorkerService extends(Service) {
}