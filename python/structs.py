
class Request():
  def __init__(self,reqid,accountid,balance,transaction,outcome):
    self.requestid=reqid
    self.accountid=accountid
    self.transaction=transaction
    self.balance=balance
    self.outcome=outcome
