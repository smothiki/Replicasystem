import sys
import random
import xml.etree.ElementTree as xmlTree
import json


class Server(process):
    def setup(ps,i,n):
        self.head=ps[0]
        self.tail=ps[n-1]

        if(i==0):
           self.type="HEAD"
           output("head")
        elif(i<n-1):
           self.type="INTERNAL"
        else:
           self.type="TAIL"
        # Have to read the account Info Map from Json
        self.accountInfo = {'1234': 20,'2345':500}
        self.ProcessedTrans = {}

    def receive(msg=('sync', self.id,accNum, bal,client,reqId,outcome,op)):
        #output("finally %r."%(selfId))
        if(accNum not in accountInfo):
             createAccount(accNum)
        accountInfo[accNum]=bal
        #output("finally updated %r."%(bal))
        sync(accNum,bal,client,reqId,outcome,op)

    def sync(accNum,bal,client,reqId,outcome,op):
        #output("sync func")
        ProcessedTrans[reqId]=[accNum,bal,op,outcome]
        #output("Processed Transactions %r"%(ProcessedTrans))
        if i<n-1:
           next= ps[i+1]
        else:
           next=0
        if(next!=0):
           send(('sync',self.id,accNum,bal,client,reqId,outcome,op), to=next)
        else:
           send(('reply',reqId,outcome,bal),to=client)


    def receive(msg=('getbalance',reqId,accNum),from_=p):
        if(accNum not in accountInfo):
             createAccount(accNum)
        balance=accountInfo[accNum]
        outcome="Processed"
        send(('reply',reqId,outcome,balance),to=p)

    def validateRequest(reqId,accNum,outcome,op):
         if(reqId not in ProcessedTrans):
             return "new"
         elif(reqId in ProcessedTrans):
             val=ProcessedTrans[reqId]
             if(val[0]==accNum and val[2]==op):
                 output(val[3])
                 return val[3]
             else:
                 output("came here")
                 return "InconsistentWithHistory"


    def receive(msg=('withdrawal',reqId,accNum,amt),from_=p):
        if(accNum not in accountInfo):
             createAccount(accNum)
        balance=accountInfo[accNum]
        outcome=''
        reqState=validateRequest(reqId,accNum,outcome,'withdrawal')
        if(reqState=="new"):
            if(balance<amt):
                outcome='Insufficient Funds'
            else:
                balance=balance-amt
                outcome='Processed'
        else:
             outcome=reqState
        sync(accNum,balance,p,reqId,outcome,'withdrawal')

    def createAccount(accNum):
        accountInfo[accNum]=0

    def receive(msg=('deposit',reqId,accNum,amt),from_=p):
        #output("receive deposit")
        if(accNum not in accountInfo):
             createAccount(accNum)
        balance=accountInfo[accNum]
        balance=balance+amt
        accountInfo[accNum]=balance
        outcome='Processed'
        sync(accNum,balance,p,reqId,outcome,'deposit')

    def main():
        # TO DO : have to stop these servers
        if(type=='HEAD'):
            await(some(received(('deposit',reqId,accNum,amt),from_=p)))
        else:
            await(some(received(('sync',self.id,accNum,bal,client,reqId,outcome,op))))

class Client(process):
    def setup(head,tail,clientId):
        self.accNum=''
        self.bankName=''
        self.balance=0
        self.reqId=0
        self.receiveQueue = set()
        self.data = ''

    def initializeMyAccDetails():
        accNum='123452'
        bankName='CitiBank'
        balance=20

    def generateReqId():
        seqNum=random.randint(0,500)
        reqId=accNum+':'+'1'+ ':'+ str(seqNum)
        output("hooo reqId.%r"%(reqId))

    def getBalance(reqId, accNum):
        #output("Balance")
        send(('getbalance',reqId,accNum,),to=tail)

    def doWithdrawal(reqId, accNum, amount):
        #output("Withdrawal")
        send(('withdrawal',reqId,accNum,amount,),to=head)

    def doDeposit(reqId, accNum, amount):
        #output("Deposit")
        send(('deposit',reqId, accNum, amount, ),to=head)

    def receive(msg=('reply',reqId,outcome,bal)):
        output("<%r,%s,%r>"%(reqId,outcome,bal))
        receiveQueue.add(reqId)

    def sendRequests():
        for i in range(0,len(data["Requests"])):
            currReq=data["Requests"][i]
            if(currReq["Operation"]=="Query"):
                getBalance(currReq["ReqId"],currReq["AccountNumber"])
                await(currReq["ReqId"] in receiveQueue)
                receiveQueue.remove(currReq["ReqId"])
            elif(currReq["Operation"]=="Withdrawal"):
                doWithdrawal(currReq["ReqId"],currReq["AccountNumber"],500)
                await(currReq["ReqId"] in receiveQueue)
                receiveQueue.remove(currReq["ReqId"])
            elif(currReq["Operation"]=="Deposit"):
                doDeposit(currReq["ReqId"],currReq["AccountNumber"],500)
                await(currReq["ReqId"] in receiveQueue)
                receiveQueue.remove(currReq["ReqId"])
            else:
                output("Sorry Requested Operation is not supported")
                #currReq["Amt"]

    def readJsonConfig():
        filename = 'client'+ str(clientId+1) + 'ReqInfo.json'
        json_data=open(filename)
        data = json.load(json_data)
        output(data["Requests"][0]["ReqId"])
        output(len(data["Requests"]))
        json_data.close()

    def main():
        readJsonConfig()
        sendRequests()
        #output("receive deposit from client")

def main():
       filename = 'configdata.xml'
       #read configuration file
       tree = xmlTree.parse(filename)
       root = tree.getroot()
       config(channel="fifo")
       #read each of the bank details from config file
       for bank in root.findall('bank'):
          bankName=bank.get('name')
          numClients=int(bank.find('./client').text)
          numServers=int(bank.find('./server').get('chain'))
          delayTime=int(bank.find('./server/delay').text)

          ps = list(new(Server, num= numServers))
          clients = new(Client,num=numClients)

          #setup(clients,[ps[0],ps[numServers-1]])

          for i, cli in enumerate(clients):
             setup(cli,[ps[0],ps[numServers-1],i])

          for i, p in enumerate(ps):
             setup(p,[ps,i,numServers])

          start(ps)
          start(clients)
          print("receive in Main")
