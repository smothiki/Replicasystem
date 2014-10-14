
import da
PatternExpr_0 = da.pat.TuplePattern([da.pat.ConstantPattern('request'), da.pat.FreePattern('request')])
PatternExpr_1 = da.pat.FreePattern('p')
PatternExpr_3 = da.pat.TuplePattern([da.pat.ConstantPattern('request'), da.pat.FreePattern('request')])
PatternExpr_4 = da.pat.FreePattern('p')
PatternExpr_5 = da.pat.TuplePattern([da.pat.ConstantPattern('reply'), da.pat.FreePattern('request')])
PatternExpr_6 = da.pat.FreePattern('p')
from structs import Request
from bank import Bank
import json

class Server(da.DistProcess):

    def __init__(self, parent, initq, channel, props):
        super().__init__(parent, initq, channel, props)
        self._ServerReceivedEvent_0 = []
        self._events.extend([da.pat.EventPattern(da.pat.ReceivedEvent, '_ServerReceivedEvent_0', PatternExpr_0, sources=[PatternExpr_1], destinations=None, timestamps=None, record_history=True, handlers=[]), da.pat.EventPattern(da.pat.ReceivedEvent, '_ServerReceivedEvent_1', PatternExpr_3, sources=[PatternExpr_4], destinations=None, timestamps=None, record_history=None, handlers=[self._Server_handler_0])])

    def setup(self, servers, index, client, name, id1):
        self.client = client
        self.name = name
        self.index = index
        self.id1 = id1
        self.servers = servers
        self.bankobj = Bank(self.name, self.id1)

    def main(self):
        request = p = None

        def ExistentialOpExpr_0():
            nonlocal request, p
            for (_, (_, _, p), (_ConstantPattern11_, request)) in self._ServerReceivedEvent_0:
                if (_ConstantPattern11_ == 'request'):
                    if True:
                        return True
            return False
        _st_label_9 = 0
        while (_st_label_9 == 0):
            _st_label_9 += 1
            if ExistentialOpExpr_0():
                _st_label_9 += 1
            else:
                super()._label('_st_label_9', block=True)
                _st_label_9 -= 1

    def _Server_handler_0(self, request, p):
        if (request.outcome in ['processed', 'inconsistent', 'insufficientfunds']):
            self.bankobj.set(request)
        elif (request.transaction == 'deposit'):
            request = self.bankobj.deposit(request)
        elif (request.transaction == 'withdraw'):
            request = self.bankobj.deposit(request)
        elif (request.transaction == 'getbalance'):
            request = self.bankobj.getbalance(request)
        if (not (self.index == 2)):
            self._send(('request', request), self.servers[(self.index + 1)])
        else:
            self.output(('<%r,%s,%r>' % (request.requestid, request.outcome, request.balance)))
            self._send(('reply', request), self.client)
    _Server_handler_0._labels = None
    _Server_handler_0._notlabels = None

class Client(da.DistProcess):

    def __init__(self, parent, initq, channel, props):
        super().__init__(parent, initq, channel, props)
        self._events.extend([da.pat.EventPattern(da.pat.ReceivedEvent, '_ClientReceivedEvent_0', PatternExpr_5, sources=[PatternExpr_6], destinations=None, timestamps=None, record_history=None, handlers=[self._Client_handler_1])])

    def setup(self, chain):
        self.chain = chain
        self.requestqueue = set()
        self.requests = []
        self.requests.append(Request('123', '1', 0, 'getbalance', 'none'))
        self.requests.append(Request('124', '2', 10, 'deposit', 'none'))
        self.requests.append(Request('124', '2', 10, 'deposit', 'none'))
        self.requests.append(Request('125', '3', 1, 'deposit', 'none'))
        self.requests.append(Request('126', '3', 2, 'withdraw', 'none'))
        self.requests.append(Request('126', '3', 3, 'deposit', 'none'))
        self.requests.append(Request('127', '4', 0, 'getbalance', 'none'))
        self.requests.append(Request('128', '5', 0, 'getbalance', 'none'))

    def main(self):
        for request in self.requests:
            self.output('inside for')
            if (request.transaction == 'getbalance'):
                self._send(('request', request), self.chain[2])
                _st_label_40 = 0
                while (_st_label_40 == 0):
                    _st_label_40 += 1
                    if (request.requestid in self.requestqueue):
                        _st_label_40 += 1
                    else:
                        super()._label('_st_label_40', block=True)
                        _st_label_40 -= 1
                else:
                    if (_st_label_40 != 2):
                        continue
                if (_st_label_40 != 2):
                    break
                self.requestqueue.remove(request.requestid)
            else:
                self._send(('request', request), self.chain[0])
                _st_label_43 = 0
                while (_st_label_43 == 0):
                    _st_label_43 += 1
                    if (request.requestid in self.requestqueue):
                        _st_label_43 += 1
                    else:
                        super()._label('_st_label_43', block=True)
                        _st_label_43 -= 1
                else:
                    if (_st_label_43 != 2):
                        continue
                if (_st_label_43 != 2):
                    break
                self.requestqueue.remove(request.requestid)

    def _Client_handler_1(self, p, request):
        self.output('sprint message')
        self.requestqueue.add(request.requestid)
        self.output(('<%r,%s,%r>' % (request.requestid, request.outcome, request.balance)))
    _Client_handler_1._labels = None
    _Client_handler_1._notlabels = None

def main():
    config = json.loads(open('../config.json').read())
    Nserver = int(config['chainlength'])
    Nclient = int(config['clients'])
    da.api.config(channel='fifo')
    servers = list(da.api.new(Server, num=Nserver))
    client = list(da.api.new(Client, num=Nclient))
    for (i, p) in enumerate(list(servers)):
        da.api.setup(p, (servers, i, client, 'wells', '123'))
    da.api.start(servers)
    da.api.setup(client, [servers])
    da.api.start(client)
