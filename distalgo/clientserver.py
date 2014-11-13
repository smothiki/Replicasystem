
import da
PatternExpr_0 = da.pat.TuplePattern([da.pat.ConstantPattern('HealthCheckmsg'), da.pat.FreePattern('serverId')])
PatternExpr_1 = da.pat.FreePattern('p')
PatternExpr_2 = da.pat.TuplePattern([da.pat.ConstantPattern('HealthCheckmsg'), da.pat.BoundPattern('_BoundPattern5_')])
PatternExpr_4 = da.pat.TuplePattern([da.pat.ConstantPattern('request'), da.pat.FreePattern('request')])
PatternExpr_5 = da.pat.FreePattern('p')
PatternExpr_7 = da.pat.TuplePattern([da.pat.ConstantPattern('ImTail'), da.pat.FreePattern('tindex')])
PatternExpr_8 = da.pat.FreePattern('p')
PatternExpr_9 = da.pat.TuplePattern([da.pat.ConstantPattern('urnext'), da.pat.FreePattern('tindex')])
PatternExpr_10 = da.pat.FreePattern('p')
PatternExpr_11 = da.pat.TuplePattern([da.pat.ConstantPattern('request'), da.pat.FreePattern('request')])
PatternExpr_12 = da.pat.FreePattern('p')
PatternExpr_13 = da.pat.TuplePattern([da.pat.ConstantPattern('ImHead'), da.pat.FreePattern('index')])
PatternExpr_14 = da.pat.FreePattern('p')
PatternExpr_15 = da.pat.TuplePattern([da.pat.ConstantPattern('ImTail'), da.pat.FreePattern('index')])
PatternExpr_16 = da.pat.FreePattern('p')
PatternExpr_17 = da.pat.TuplePattern([da.pat.ConstantPattern('reply'), da.pat.FreePattern('request')])
PatternExpr_18 = da.pat.FreePattern('p')
from structs import Request
from structs import Requests
from bank import Bank
import json
import sys
import time
import _thread

class Master(da.DistProcess):

    def __init__(self, parent, initq, channel, props):
        super().__init__(parent, initq, channel, props)
        self._MasterReceivedEvent_1 = []
        self._events.extend([da.pat.EventPattern(da.pat.ReceivedEvent, '_MasterReceivedEvent_0', PatternExpr_0, sources=[PatternExpr_1], destinations=None, timestamps=None, record_history=None, handlers=[self._Master_handler_0]), da.pat.EventPattern(da.pat.ReceivedEvent, '_MasterReceivedEvent_1', PatternExpr_2, sources=None, destinations=None, timestamps=None, record_history=True, handlers=[])])

    def setup(self, servers, client, Nlength):
        self.Nlength = Nlength
        self.servers = servers
        self.client = client
        self.serverisAlive = {}
        self.head = 0
        self.tail = (self.Nlength - 1)
        for (i, s) in enumerate(list(self.servers)):
            self.serverisAlive[i] = 0

    def main(self):
        self.output(('%r.' % self.serverisAlive))
        _thread.start_new_thread(self.checkHealthmsg, ('Thread', 20))

        def ExistentialOpExpr_0():
            for (_, _, (_ConstantPattern14_, _BoundPattern15_)) in self._MasterReceivedEvent_1:
                if (_ConstantPattern14_ == 'HealthCheckmsg'):
                    if (_BoundPattern15_ == self.id):
                        if True:
                            return True
            return False
        _st_label_38 = 0
        while (_st_label_38 == 0):
            _st_label_38 += 1
            if ExistentialOpExpr_0():
                _st_label_38 += 1
            else:
                super()._label('_st_label_38', block=True)
                _st_label_38 -= 1

    def checkHealthmsg(self, threadName, delay):
        while True:
            time.sleep(delay)
            self.output(('moon %r.' % self.serverisAlive))
            for (i, s) in enumerate(list(self.servers)):
                if (self.serverisAlive[i] == 0):
                    if (i == self.head):
                        self.output('here1')
                        newHead = (i + 1)
                        self._send(('ImHead', newHead), self.client)
                    elif (i == self.tail):
                        self.output('here2')
                        newTail = (i - 1)
                        self._send(('ImTail', newTail), s[(i - 2)])
                        self._send(('ImTail', newTail), self.client)
                    else:
                        self._send(('urnext', (i + 1)), s[(i - 1)])

    def _Master_handler_0(self, serverId, p):
        self.output(('*****%r.' % serverId))
        self.serverisAlive[serverId] = 1
    _Master_handler_0._labels = None
    _Master_handler_0._notlabels = None

class Server(da.DistProcess):

    def __init__(self, parent, initq, channel, props):
        super().__init__(parent, initq, channel, props)
        self._ServerReceivedEvent_0 = []
        self._events.extend([da.pat.EventPattern(da.pat.ReceivedEvent, '_ServerReceivedEvent_0', PatternExpr_4, sources=[PatternExpr_5], destinations=None, timestamps=None, record_history=True, handlers=[]), da.pat.EventPattern(da.pat.ReceivedEvent, '_ServerReceivedEvent_1', PatternExpr_7, sources=[PatternExpr_8], destinations=None, timestamps=None, record_history=None, handlers=[self._Server_handler_1]), da.pat.EventPattern(da.pat.ReceivedEvent, '_ServerReceivedEvent_2', PatternExpr_9, sources=[PatternExpr_10], destinations=None, timestamps=None, record_history=None, handlers=[self._Server_handler_2]), da.pat.EventPattern(da.pat.ReceivedEvent, '_ServerReceivedEvent_3', PatternExpr_11, sources=[PatternExpr_12], destinations=None, timestamps=None, record_history=None, handlers=[self._Server_handler_3])])

    def setup(self, servers, index, client, Nlength, Nchain, master):
        self.index = index
        self.Nchain = Nchain
        self.master = master
        self.servers = servers
        self.client = client
        self.Nlength = Nlength
        self.bankobj = Bank('wells', '123')
        tempindex = self.index
        self.startindex = 0
        self.head = self.servers[0]
        self.next = self.servers[(self.index + 1)]
        self.endindex = ((self.Nlength * self.Nchain) - 1)
        self.output(str(self.index))

    def main(self):
        _thread.start_new_thread(self.sendHealthCheck, ('Thread', 5, self.index, self.master))
        p = request = None

        def ExistentialOpExpr_1():
            nonlocal p, request
            for (_, (_, _, p), (_ConstantPattern27_, request)) in self._ServerReceivedEvent_0:
                if (_ConstantPattern27_ == 'request'):
                    if True:
                        return True
            return False
        _st_label_50 = 0
        while (_st_label_50 == 0):
            _st_label_50 += 1
            if ExistentialOpExpr_1():
                _st_label_50 += 1
            else:
                super()._label('_st_label_50', block=True)
                _st_label_50 -= 1
        _st_label_51 = 0
        while (_st_label_51 == 0):
            _st_label_51 += 1
            if False:
                _st_label_51 += 1
            else:
                super()._label('_st_label_51', block=True)
                _st_label_51 -= 1

    def sendHealthCheck(self, threadName, delay, index, master):
        self.output('inside sendHealthCheck')
        self.output((' %r.' % index))
        while True:
            time.sleep(delay)
            self._send(('HealthCheckmsg', index), master)

    def _Server_handler_1(self, tindex, p):
        self.endindex = tindex
    _Server_handler_1._labels = None
    _Server_handler_1._notlabels = None

    def _Server_handler_2(self, tindex, p):
        self.next = self.servers[tindex]
    _Server_handler_2._labels = None
    _Server_handler_2._notlabels = None

    def _Server_handler_3(self, request, p):
        if (request.outcome in ['processed', 'inconsistent', 'insufficientfunds']):
            self.bankobj.set(request)
        elif (request.transaction == 'deposit'):
            request = self.bankobj.deposit(request)
        elif (request.transaction == 'withdraw'):
            request = self.bankobj.withdraw(request)
        elif (request.transaction == 'getbalance'):
            request = self.bankobj.getbalance(request)
        if (self.index < self.endindex):
            self._send(('request', request), self.next)
        else:
            self._send(('reply', request), self.client)
    _Server_handler_3._labels = None
    _Server_handler_3._notlabels = None

class Client(da.DistProcess):

    def __init__(self, parent, initq, channel, props):
        super().__init__(parent, initq, channel, props)
        self._events.extend([da.pat.EventPattern(da.pat.ReceivedEvent, '_ClientReceivedEvent_0', PatternExpr_13, sources=[PatternExpr_14], destinations=None, timestamps=None, record_history=None, handlers=[self._Client_handler_4]), da.pat.EventPattern(da.pat.ReceivedEvent, '_ClientReceivedEvent_1', PatternExpr_15, sources=[PatternExpr_16], destinations=None, timestamps=None, record_history=None, handlers=[self._Client_handler_5]), da.pat.EventPattern(da.pat.ReceivedEvent, '_ClientReceivedEvent_2', PatternExpr_17, sources=[PatternExpr_18], destinations=None, timestamps=None, record_history=None, handlers=[self._Client_handler_6])])

    def setup(self, chain, master, Nlength):
        self.master = master
        self.Nlength = Nlength
        self.chain = chain
        self.waitfor = set()
        self.requests = Requests(3, 'getbalance')
        self.requests.getRequestList()
        self.head = self.chain[0]
        self.tail = self.chain[(self.Nlength - 1)]
        self.done = False

    def main(self):
        for request in self.requests.requests:
            if (request.requestid in self.waitfor):
                self.waitfor.remove(request.requestid)
            self.waitfor.add(request.requestid)
            if (request.transaction == 'getbalance'):
                self._send(('request', request), self.tail)
            else:
                self._send(('request', request), self.head)
        _st_label_90 = 0
        self._timer_start()
        while (_st_label_90 == 0):
            _st_label_90 += 1
            if False:
                pass
                _st_label_90 += 1
            elif self._timer_expired:
                self.output('timeout')
                _st_label_90 += 1
            else:
                super()._label('_st_label_90', block=True, timeout=20)
                _st_label_90 -= 1
        self.output('finished request')

    def _Client_handler_4(self, p, index):
        self.head = self.chain[index]
    _Client_handler_4._labels = None
    _Client_handler_4._notlabels = None

    def _Client_handler_5(self, p, index):
        self.tail = self.chain[index]
    _Client_handler_5._labels = None
    _Client_handler_5._notlabels = None

    def _Client_handler_6(self, p, request):
        if (request.requestid in self.waitfor):
            self.waitfor.remove(request.requestid)
        self.output('client')
        self.output(('<%r,%s,%r>' % (request.requestid, request.outcome, request.balance)))
        if (len(self.waitfor) == 0):
            self.done = True
    _Client_handler_6._labels = None
    _Client_handler_6._notlabels = None

def main():
    configs = json.loads(open('/Users/ram/deistests/src/github.com/replicasystem/config/config.json').read())
    Nserver = int(configs['chainlength'])
    timeout = 1
    da.api.config(channel='fifo')
    servers = list(da.api.new(Server, num=Nserver))
    client = da.api.new(Client, num=1)
    master = da.api.new(Master, num=1)
    for (i, p) in enumerate(list(servers)):
        da.api.setup(p, (servers, i, client, master))
    da.api.start(servers)
    da.api.setup(client, [servers, master, Nserver])
    da.api.setup(master, [servers, client, Nserver])
    da.api.start(client)
