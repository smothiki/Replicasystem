from structs import Request

# Bank class that handles bank system

class Bank():
  def __init__(self,name,id):
    self.name = name
    self.bankid=id
    self.amap={}
    self.tmap={}

  def recordtransaction(self,reqid,typet):
    self.tmap[reqid]=typet

  def checktransaction(self,reqid,typet):
    if reqid in self.tmap:
      if self.tmap[reqid]==typet:
        return "processed"
      else :
        return "inconsistent"
    return "new"

  def checkaccountid(self,account):
    if account not in self.amap:
      self.amap[account]=0


  def getbalance(self,request):
    self.checkaccountid(request.accountid)
    return Request(request.requestid,request.accountid,self.amap[request.accountid],"getbalance","processed")


  def deposit(self,request):
    self.checkaccountid(request.accountid)
    resp=self.checktransaction(request.requestid,"deposit")
    if resp == "new" :
      resp = "processed"
      bal=self.amap[request.accountid]
      bal =bal + request.balance
      self.amap[request.accountid]=bal
      self.tmap[request.requestid]="deposit"
    return Request(request.requestid,request.accountid,self.amap[request.accountid],"deposit",resp)

  def withdraw(self,request):
    self.checkaccountid(request.accountid)
    resp=self.checktransaction(request.requestid,"withdraw")
    if resp == "new":
      resp="processed"
      bal=self.amap[request.accountid]
      if bal-request.balance < 0 :
        self.tmap[request.requestid]="withdraw"
        return Request(request.requestid,request.accountid,self.amap[request.accountid],"withdraw","insufficientfunds")
      bal =bal - request.balance
      self.amap[request.accountid]=bal
      self.tmap[request.requestid]="withdraw"
    return Request(request.requestid,request.accountid,self.amap[request.accountid],"withdraw",resp)

  def set(self,request):
    self.checkaccountid(request.accountid)
    self.amap[request.accountid]=request.balance
    self.tmap[request.requestid]=request.transaction


# bankobj=Bank("wells","wells")
# requests=[]
# requests.append(Request("123","1",0,"getbalance","none"))
# requests.append(Request("124","2",10,"deposit","none"))
# requests.append(Request("124","2",10,"deposit","none"))
# requests.append(Request("125","3",1,"deposit","none"))
# requests.append(Request("126","3",2,"withdraw","none"))
# requests.append(Request("126","3",3,"deposit","none"))
# requests.append(Request("127","4",0,"getbalance","none"))
# requests.append(Request("128","5",0,"getbalance","none"))
# for request in requests:
#   if request.transaction == "getbalance" :
#     reply=bankobj.getbalance(request)
#   elif request.transaction == "withdraw":
#     reply=bankobj.withdraw(request)
#   elif request.transaction == "deposit":
#     reply=bankobj.deposit(request)
#   print(reply.requestid,reply.outcome,reply.balance)
