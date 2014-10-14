import json
import uuid


# 7d529dd4-548b-4258-aa8e-23e34dc8d43d



class Request():
  def __init__(self,reqid,accountid,balance,transaction,outcome):
    self.requestid=reqid
    self.accountid=accountid
    self.transaction=transaction
    self.balance=balance
    self.outcome=outcome

def Randomreq(balance,typet):
  uid=str(uuid.uuid4())
  return Request(uid[0:4],uid[4:8],balance,typet,"none")

class Requests():
  def __init__(self,prob,typet):
    self.prob=prob
    self.typet=typet
    self.requests=[]

  def getRequestTypes(self,counts,transaction):
    requestJson = json.loads(open("/Users/ram/deistests/src/github.com/replicasystem/config/request.json").read())
    Reqs = requestJson["requests"]
    getReqs = Reqs[transaction]
    if len(getReqs)-counts <0 :
      for i in range(counts-len(getReqs)):
        self.requests.append(Randomreq(0,transaction))
      for req in getReqs:
        self.requests.append(Request(req["requestid"],req["account"],int(req["balance"]),req["transaction"],req["outcome"]))
    else:
      for i in range(counts):
        self.requests.append(Request(getReqs[i]["requestid"],getReqs[i]["account"],int(getReqs[i]["balance"]),getReqs[i]["transaction"],getReqs[i]["outcome"]))

  def getRequestList(self):
    configs = json.loads(open("/Users/ram/deistests/src/github.com/replicasystem/config/config.json").read())
    MaxReqs = int(configs["MaxRequests"])
    print(MaxReqs)
    types = ["getbalance", "deposit", "withdraw"]
    if self.prob ==0 :
      for i in range(3):
        self.getRequestTypes(6,types[i])
      self.requests.append(Randomreq(0,"getbalance"))
      self.requests.append(Randomreq(0,"deposit"))
    else:
      rem = MaxReqs - (2*self.prob)
      rem =int(rem/2)
      self.prob=self.prob*2
      self.getRequestTypes(2*self.prob,self.typet)
      for i in range(3):
        if self.typet != types[i]:
          self.getRequestTypes(rem,types[i])
