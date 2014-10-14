from structs import Request
from bank import Bank

class Server(process):
   def setup(servers,index,client,name,id1):
     self.bankobj=Bank(name,id1)

   def main():
     while True:
       await(some(received(("request",request),from_=p)))


   # when receiving request from others, enque and reply
   def receive(msg=("request",request),from_=p):
     if request.outcome in ["processed","inconsistent","insufficientfunds"]:
       self.bankobj.set(request)
     elif request.transaction == "deposit":
       request=self.bankobj.deposit(request)
     elif request.transaction == "withdraw":
       request=self.bankobj.deposit(request)
     elif request.transaction == "getbalance":
       request=self.bankobj.getbalance(request)
     output(request.outcome)
     if index != 2:
       output("sending to next server")
       send(("request",request,),to=servers[index+1])
     else :
       output("sending to client")
       send(("reply",request),to=client)


class Client(process):
   def setup(chain):
     self.requests=[]
     requests.append(Request("123","1",0,"getbalance","none"))
     requests.append(Request("124","2",10,"deposit","none"))
     requests.append(Request("124","2",10,"deposit","none"))
     requests.append(Request("125","3",1,"deposit","none"))
     requests.append(Request("126","3",2,"withdraw","none"))
     requests.append(Request("126","3",3,"deposit","none"))
     requests.append(Request("127","4",0,"getbalance","none"))
     requests.append(Request("128","5",0,"getbalance","none"))

   def main():
     for request in requests:
       if request.transaction == "getbalance" :
         send(("request",request),to=chain[2])
       else:
         send(("request",request),to=chain[0])
     while True:
       await(some(received(("reply",request),from_=p)))

   def receive(msg=('reply',request),from_=p):
     output("sprint message")
     output(request.outcome)


def main():
   servers1 = list(new(Server, num=3))
   client1 = list(new(Client, num=1))
   for i,p in enumerate(list(servers1)):
     setup(p,(servers1,i,client1,"wells","123"))
   start(servers1)
   setup(client1, [servers1])
   start(client1)
  #  servers2 = list(new(Server, num=3))
  #  client2 = list(new(Client, num=1))
  #  for i,p in enumerate(list(servers2)):
  #    setup(p,(servers2,i,client2,"wells","123"))
  #  start(servers2)
  #  setup(client2, [servers2])
  #  start(client2)
