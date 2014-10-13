
import da
PatternExpr_0 = da.pat.TuplePattern([da.pat.ConstantPattern('request'), da.pat.FreePattern('request')])
PatternExpr_1 = da.pat.FreePattern('p')
PatternExpr_3 = da.pat.TuplePattern([da.pat.ConstantPattern('request'), da.pat.FreePattern('request')])
PatternExpr_4 = da.pat.FreePattern('p')
PatternExpr_5 = da.pat.TuplePattern([da.pat.ConstantPattern('reply'), da.pat.FreePattern('request')])
PatternExpr_6 = da.pat.FreePattern('p')
PatternExpr_8 = da.pat.TuplePattern([da.pat.ConstantPattern('reply'), da.pat.FreePattern('request')])
PatternExpr_9 = da.pat.FreePattern('p')
from structs import Request
from bank import Bank

class Server(da.DistProcess):

    def __init__(self, parent, initq, channel, props):
        super().__init__(parent, initq, channel, props)
        self._ServerReceivedEvent_0 = []
        self._events.extend([da.pat.EventPattern(da.pat.ReceivedEvent, '_ServerReceivedEvent_0', PatternExpr_0, sources=[PatternExpr_1], destinations=None, timestamps=None, record_history=True, handlers=[]), da.pat.EventPattern(da.pat.ReceivedEvent, '_ServerReceivedEvent_1', PatternExpr_3, sources=[PatternExpr_4], destinations=None, timestamps=None, record_history=None, handlers=[self._Server_handler_0])])

    def setup(self, servers, serve, client, name, id1):
        self.serve = serve
        self.client = client
        self.servers = servers
        self.name = name
        self.id1 = id1
        self.bankobj = Bank(self.name, self.id1)

    def main(self):
        p = request = None

        def ExistentialOpExpr_0():
            nonlocal p, request
            for (_, (_, _, p), (_ConstantPattern11_, request)) in self._ServerReceivedEvent_0:
                if (_ConstantPattern11_ == 'request'):
                    if True:
                        return True
            return False
        _st_label_8 = 0
        while (_st_label_8 == 0):
            _st_label_8 += 1
            if ExistentialOpExpr_0():
                _st_label_8 += 1
            else:
                super()._label('_st_label_8', block=True)
                _st_label_8 -= 1

    def _Server_handler_0(self, p, request):
        if (request.outcome in ['processed', 'inconsistent', 'insufficientfunds']):
            self.bankobj.set(request)
        elif (request.transaction == 'deposit'):
            request = self.bankobj.deposit(request)
        elif (request.transaction == 'withdraw'):
            request = self.bankobj.deposit(request)
        elif (request.transaction == 'getbalance'):
            request = self.bankobj.getbalance(request)
        self.output(request.outcome)
        if (not (self.serve == 4)):
            self._send(('request', request), self.servers[(self.serve + 1)])
        else:
            self.output('sending to client')
            self._send(('reply', request), self.client)
            self.output('sent to client')
    _Server_handler_0._labels = None
    _Server_handler_0._notlabels = None

class Client(da.DistProcess):

    def __init__(self, parent, initq, channel, props):
        super().__init__(parent, initq, channel, props)
        self._ClientReceivedEvent_0 = []
        self._events.extend([da.pat.EventPattern(da.pat.ReceivedEvent, '_ClientReceivedEvent_0', PatternExpr_5, sources=[PatternExpr_6], destinations=None, timestamps=None, record_history=True, handlers=[]), da.pat.EventPattern(da.pat.ReceivedEvent, '_ClientReceivedEvent_1', PatternExpr_8, sources=[PatternExpr_9], destinations=None, timestamps=None, record_history=None, handlers=[self._Client_handler_1])])

    def setup(self, chain):
        self.chain = chain
        self.request = Request('1234', '1234', 12, 'deposit', 'none')

    def main(self):
        self._send(('request', self.request), self.chain[0])
        self.output('sending to server')
        request = p = None

        def ExistentialOpExpr_1():
            nonlocal request, p
            for (_, (_, _, p), (_ConstantPattern28_, request)) in self._ClientReceivedEvent_0:
                if (_ConstantPattern28_ == 'reply'):
                    if True:
                        return True
            return False
        _st_label_30 = 0
        while (_st_label_30 == 0):
            _st_label_30 += 1
            if ExistentialOpExpr_1():
                _st_label_30 += 1
            else:
                super()._label('_st_label_30', block=True)
                _st_label_30 -= 1
        self.output('finished request')

    def _Client_handler_1(self, request, p):
        self.output('sprint message')
        self.output(request.outcome)
    _Client_handler_1._labels = None
    _Client_handler_1._notlabels = None

def main():
    nprocs = 4
    servers = list(da.api.new(Server, num=5))
    client = list(da.api.new(Client, num=1))
    for (i, p) in enumerate(list(servers)):
        da.api.setup(p, (servers, i, client, 'wells', '123'))
    da.api.start(servers)
    da.api.setup(client, [servers])
    da.api.start(client)
