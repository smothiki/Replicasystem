from structs import Request

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
      if bal-amount < 0 :
        return Request(request.requestid,request.accountid,self.amap[request.accountid],"withdraw","insufficientfunds")
      bal =bal - amount
      self.amap[request.accountid]=bal
      self.tmap[request.requestid]="withdraw"
    return Request(request.requestid,request.accountid,self.amap[request.accountid],"withdraw",resp)

  def set(self,request):
    self.checkaccountid(request.accountid)
    self.amap[request.accountid]=request.balance
    self.tmap[request.requestid]=request.transaction
