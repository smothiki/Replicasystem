
import da
PatternExpr_0 = da.pat.TuplePattern([da.pat.ConstantPattern('HealthCheckmsg'), da.pat.FreePattern('serverId')])
PatternExpr_1 = da.pat.FreePattern('p')
PatternExpr_2 = da.pat.TuplePattern([da.pat.ConstantPattern('request'), da.pat.FreePattern('request')])
PatternExpr_3 = da.pat.FreePattern('p')
PatternExpr_5 = da.pat.TuplePattern([da.pat.ConstantPattern('ImTail'), da.pat.FreePattern('tindex')])
PatternExpr_6 = da.pat.FreePattern('p')
PatternExpr_7 = da.pat.TuplePattern([da.pat.ConstantPattern('urnext'), da.pat.FreePattern('tindex')])
PatternExpr_8 = da.pat.FreePattern('p')
PatternExpr_9 = da.pat.TuplePattern([da.pat.ConstantPattern('request'), da.pat.FreePattern('request')])
PatternExpr_10 = da.pat.FreePattern('p')
PatternExpr_11 = da.pat.TuplePattern([da.pat.ConstantPattern('ImHead'), da.pat.FreePattern('index')])
PatternExpr_12 = da.pat.FreePattern('p')
PatternExpr_13 = da.pat.TuplePattern([da.pat.ConstantPattern('ImTail'), da.pat.FreePattern('index')])
PatternExpr_14 = da.pat.FreePattern('p')
PatternExpr_15 = da.pat.TuplePattern([da.pat.ConstantPattern('reply'), da.pat.FreePattern('request')])
PatternExpr_16 = da.pat.FreePattern('p')
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
        self._events.extend([da.pat.EventPattern(da.pat.ReceivedEvent, '_MasterReceivedEvent_0', PatternExpr_0, sources=[PatternExpr_1], destinations=None, timestamps=None, record_history=None, handlers=[self._Master_handler_0])])

    def setup(self, servers, client, Nlength):
        self.client = client
        self.servers = servers
        self.Nlength = Nlength
        self.serverisAlive = {}
        self.head = 0
        self.tail = (self.Nlength - 1)
        for (i, s) in enumerate(list(self.servers)):
            self.serverisAlive[i] = 1

    def main(self):
        self.output(('%r.' % self.serverisAlive))
        _thread.start_new_thread(self.checkHealthmsg, ('Thread', 6))
        _thread.start_new_thread(self.messuphead, ('Thread1', 4))
        _thread.start_new_thread(self.messuptail, ('Thread2', 8))
        _thread.start_new_thread(self.messupserver, ('Thread3', 1, 6))
        _st_label_49 = 0
        while (_st_label_49 == 0):
            _st_label_49 += 1
            if False:
                _st_label_49 += 1
            else:
                super()._label('_st_label_49', block=True)
                _st_label_49 -= 1

    def checkHealthmsg(self, threadName, delay):
        while True:
            time.sleep(1)
            self.output(('state %r.' % self.serverisAlive))
            for (i, s) in enumerate(self.servers):
                if (self.serverisAlive[i] == 0):
                    if (i == self.head):
                        self.output('killing head')
                        newHead = (i + 1)
                        self._send(('ImHead', newHead), self.client)
                    elif (i == self.tail):
                        self.output('killing tail')
                        newTail = (i - 1)
                        self._send(('ImTail', newTail), self.servers[(i - 2)])
                        self._send(('ImTail', newTail), self.client)
                    else:
                        self._send(('urnext', (i + 1)), self.servers[(i - 1)])

    def messuphead(self, threadName, headkill):
        time.sleep(headkill)
        self.serverisAlive[0] = 0

    def messuptail(self, threadName, tailkill):
        time.sleep(tailkill)
        self.serverisAlive[(self.Nlength - 1)] = 0

    def messupserver(self, threadName, index, timetokill):
        time.sleep(timetokill)
        self.serverisAlive[index] = 0

    def _Master_handler_0(self, p, serverId):
        self.serverisAlive[serverId] = 1
    _Master_handler_0._labels = None
    _Master_handler_0._notlabels = None

class Server(da.DistProcess):

    def __init__(self, parent, initq, channel, props):
        super().__init__(parent, initq, channel, props)
        self._ServerReceivedEvent_0 = []
        self._events.extend([da.pat.EventPattern(da.pat.ReceivedEvent, '_ServerReceivedEvent_0', PatternExpr_2, sources=[PatternExpr_3], destinations=None, timestamps=None, record_history=True, handlers=[]), da.pat.EventPattern(da.pat.ReceivedEvent, '_ServerReceivedEvent_1', PatternExpr_5, sources=[PatternExpr_6], destinations=None, timestamps=None, record_history=None, handlers=[self._Server_handler_1]), da.pat.EventPattern(da.pat.ReceivedEvent, '_ServerReceivedEvent_2', PatternExpr_7, sources=[PatternExpr_8], destinations=None, timestamps=None, record_history=None, handlers=[self._Server_handler_2]), da.pat.EventPattern(da.pat.ReceivedEvent, '_ServerReceivedEvent_3', PatternExpr_9, sources=[PatternExpr_10], destinations=None, timestamps=None, record_history=None, handlers=[self._Server_handler_3])])

    def setup(self, servers, index, client, Nlength, master):
        self.client = client
        self.servers = servers
        self.index = index
        self.Nlength = Nlength
        self.master = master
        self.bankobj = Bank('wells', '123')
        self.startindex = 0
        self.head = self.servers[0]
        if (self.index == (self.Nlength - 1)):
            self.output(str(self.index))
            self.next = self.client
        else:
            self.next = self.servers[(self.index + 1)]
        self.endindex = (self.Nlength - 1)
        self.output(('serverindex' + str(self.index)))

    def main(self):
        _thread.start_new_thread(self.sendHealthCheck, ('Thread', 3, self.index, self.master))
        p = request = None

        def ExistentialOpExpr_0():
            nonlocal p, request
            for (_, (_, _, p), (_ConstantPattern15_, request)) in self._ServerReceivedEvent_0:
                if (_ConstantPattern15_ == 'request'):
                    if True:
                        return True
            return False
        _st_label_63 = 0
        while (_st_label_63 == 0):
            _st_label_63 += 1
            if ExistentialOpExpr_0():
                _st_label_63 += 1
            else:
                super()._label('_st_label_63', block=True)
                _st_label_63 -= 1
        _st_label_64 = 0
        while (_st_label_64 == 0):
            _st_label_64 += 1
            if False:
                _st_label_64 += 1
            else:
                super()._label('_st_label_64', block=True)
                _st_label_64 -= 1

    def sendHealthCheck(self, threadName, delay, index, master):
        self.output('inside sendHealthCheck')
        self.output((' %r.' % index))
        while True:
            time.sleep(delay)
            self._send(('HealthCheckmsg', index), master)

    def _Server_handler_1(self, tindex, p):
        self.output('updated tail server end')
        self.endindex = tindex
    _Server_handler_1._labels = None
    _Server_handler_1._notlabels = None

    def _Server_handler_2(self, tindex, p):
        self.output('updated next server end')
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
            self.output(('<%r,%s,%r>' % (request.requestid, request.outcome, request.balance)))
        else:
            self._send(('reply', request), self.client)
    _Server_handler_3._labels = None
    _Server_handler_3._notlabels = None

class Client(da.DistProcess):

    def __init__(self, parent, initq, channel, props):
        super().__init__(parent, initq, channel, props)
        self._events.extend([da.pat.EventPattern(da.pat.ReceivedEvent, '_ClientReceivedEvent_0', PatternExpr_11, sources=[PatternExpr_12], destinations=None, timestamps=None, record_history=None, handlers=[self._Client_handler_4]), da.pat.EventPattern(da.pat.ReceivedEvent, '_ClientReceivedEvent_1', PatternExpr_13, sources=[PatternExpr_14], destinations=None, timestamps=None, record_history=None, handlers=[self._Client_handler_5]), da.pat.EventPattern(da.pat.ReceivedEvent, '_ClientReceivedEvent_2', PatternExpr_15, sources=[PatternExpr_16], destinations=None, timestamps=None, record_history=None, handlers=[self._Client_handler_6])])

    def setup(self, chain, master, Nlength):
        self.Nlength = Nlength
        self.chain = chain
        self.master = master
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
            time.sleep(1)
            if (request.transaction == 'getbalance'):
                self._send(('request', request), self.tail)
            else:
                self._send(('request', request), self.head)
        _st_label_107 = 0
        self._timer_start()
        while (_st_label_107 == 0):
            _st_label_107 += 1
            if False:
                pass
                _st_label_107 += 1
            elif self._timer_expired:
                self.output('timeout')
                _st_label_107 += 1
            else:
                super()._label('_st_label_107', block=True, timeout=40)
                _st_label_107 -= 1
        self.output('finished request')

    def _Client_handler_4(self, index, p):
        self.output('new head updated')
        self.head = self.chain[index]
    _Client_handler_4._labels = None
    _Client_handler_4._notlabels = None

    def _Client_handler_5(self, p, index):
        self.output('new head updated')
        self.tail = self.chain[index]
    _Client_handler_5._labels = None
    _Client_handler_5._notlabels = None

    def _Client_handler_6(self, request, p):
        if (request.requestid in self.waitfor):
            self.waitfor.remove(request.requestid)
        self.output('client side reply')
        self.output(('<%r,%s,%r>' % (request.requestid, request.outcome, request.balance)))
        if (len(self.waitfor) == 0):
            self.done = True
    _Client_handler_6._labels = None
    _Client_handler_6._notlabels = None

def main():
    configs = json.loads(open('/Users/ram/deistests/src/github.com/replicasystem/config/config01.json').read())
    Nserver = int(configs['chainlength'])
    timeout = 1
    da.api.config(channel='fifo')
    servers = list(da.api.new(Server, num=Nserver))
    client = da.api.new(Client, num=1)
    master = da.api.new(Master, num=1)
    for (i, p) in enumerate(list(servers)):
        da.api.setup(p, (servers, i, client, Nserver, master))
    da.api.start(servers)
    da.api.setup(client, [servers, master, Nserver])
    da.api.setup(master, [servers, client, Nserver])
    da.api.start(client)
    da.api.start(master)
