#!/usr/bin/python

"""
consoles.py: bring up a bunch of miniature consoles on a virtual network

This demo shows how to monitor a set of nodes by using
Node's monitor() and Tkinter's createfilehandler().

We monitor nodes in a couple of ways:

- First, each individual node is monitored, and its output is added
  to its console window

- Second, each time a console window gets iperf output, it is parsed
  and accumulated. Once we have output for all consoles, a bar is
  added to the bandwidth graph.

The consoles also support limited interaction:

- Pressing "return" in a console will send a command to it

- Pressing the console's title button will open up an xterm

Bob Lantz, April 2010

"""

import re

from Tkinter import Frame, Button, Label, Text, Scrollbar, Canvas, Wm, READABLE

from mininet.log import setLogLevel
from mininet.topolib import TreeNet
from mininet.term import makeTerms, cleanUpScreens
from mininet.util import quietRun
import time, threading
from random import randint


class Console( Frame ):
    "A simple console on a host."

    def __init__( self, parent, net, node, height=30, width=64, title='Node' ):
        Frame.__init__( self, parent )

        self.net = net
        self.node = node
        self.prompt = node.name + '# '
        self.height, self.width, self.title = height, width, title

        # Initialize widget styles
        self.buttonStyle = { 'font': 'Monaco 7' }
        self.textStyle = {
            'font': 'Monaco 7',
            'bg': 'black',
            'fg': 'green',
            'width': self.width,
            'height': self.height,
            'relief': 'sunken',
            'insertbackground': 'green',
            'highlightcolor': 'green',
            'selectforeground': 'black',
            'selectbackground': 'green'
        }

        # Set up widgets
        self.text = self.makeWidgets( )
        self.bindEvents()
        # && export TERM=dumb
        tp, num = node.name[0], node.name[1:]
        print tp + ' ' + num
        addedCmd = ''
        if tp == 'h':
           addedCmd = 'go run main.go '+node.IP()+' '+'8000'
           if num != '1':
               addedCmd = addedCmd+' 10.0.0.1 8000'

           
        self.sendCmd( 'export PATH=$PATH:/usr/local/go/bin && export GOPATH=/home/kid/GoProjects/ && cd .. && export TERM=dumb && '+ addedCmd + ' &' )
        time.sleep(0.5)
        self.outputHook = None

    def makeWidgets( self ):
        "Make a label, a text area, and a scroll bar."

        def newTerm( net=self.net, node=self.node, title=self.title ):
            "Pop up a new terminal window for a node."
            net.terms += makeTerms( [ node ], title )
        label = Button( self, text=self.node.name, command=newTerm,
                        **self.buttonStyle )
        label.pack( side='top', fill='x' )
        text = Text( self, wrap='word', **self.textStyle )
        ybar = Scrollbar( self, orient='vertical', width=7,
                          command=text.yview )
        text.configure( yscrollcommand=ybar.set )
        text.pack( side='left', expand=True, fill='both' )
        ybar.pack( side='right', fill='y' )
        return text

    def bindEvents( self ):
        "Bind keyboard and file events."
        # The text widget handles regular key presses, but we
        # use special handlers for the following:
        self.text.bind( '<Return>', self.handleReturn )
        self.text.bind( '<Control-c>', self.handleInt )
        self.text.bind( '<KeyPress>', self.handleKey )
        # This is not well-documented, but it is the correct
        # way to trigger a file event handler from Tk's
        # event loop!
        self.tk.createfilehandler( self.node.stdout, READABLE,
                                   self.handleReadable )

    # We're not a terminal (yet?), so we ignore the following
    # control characters other than [\b\n\r]
    ignoreChars = re.compile( r'[\x00-\x07\x09\x0b\x0c\x0e-\x1f]+' )

    def append( self, text ):
        "Append something to our text frame."
        text = self.ignoreChars.sub( '', text )
        self.text.insert( 'end', text )
        self.text.mark_set( 'insert', 'end' )
        self.text.see( 'insert' )
        outputHook = lambda x, y: True  # make pylint happier
        if self.outputHook:
            outputHook = self.outputHook
        outputHook( self, text )

    def handleKey( self, event ):
        "If it's an interactive command, send it to the node."
        char = event.char
        if self.node.waiting:
            self.node.write( char )

    def handleReturn( self, event ):
        "Handle a carriage return."
        cmd = self.text.get( 'insert linestart', 'insert lineend' )
        # Send it immediately, if "interactive" command
        if self.node.waiting:
            self.node.write( event.char )
            return
        # Otherwise send the whole line to the shell
        pos = cmd.find( self.prompt )
        if pos >= 0:
            cmd = cmd[ pos + len( self.prompt ): ]
        self.sendCmd( cmd )

    # Callback ignores event
    def handleInt( self, _event=None ):
        "Handle control-c."
        self.node.sendInt()

    def sendCmd( self, cmd ):
        "Send a command to our node."
        if not self.node.waiting:
            print 'sending ' + cmd
            self.node.sendCmd( cmd )

    def handleReadable( self, _fds, timeoutms=None ):
        "Handle file readable event."
        data = self.node.monitor( timeoutms )
        self.append( data )
        if not self.node.waiting:
            # Print prompt
            self.append( self.prompt )

    def waiting( self ):
        "Are we waiting for output?"
        return self.node.waiting

    def waitOutput( self ):
        "Wait for any remaining output."
        while self.node.waiting:
            # A bit of a trade-off here...
            self.handleReadable( self, timeoutms=1000)
            self.update()

    def clear( self ):
        "Clear all of our text."
        self.text.delete( '1.0', 'end' )


class ConsoleApp( Frame ):

    "Simple Tk consoles for Mininet."

    menuStyle = { 'font': 'Geneva 7 bold' }

    def __init__( self, net, parent=None, width=4 ):
        Frame.__init__( self, parent )
        self.top = self.winfo_toplevel()
        self.top.title( 'Mininet' )
        self.net = net
        self.menubar = self.createMenuBar()
        cframe = self.cframe = Frame( self )
        self.consoles = {}  # consoles themselves
        self.disconnectedLinks = []
        self.disonnectThread = None
        titles = {
            'hosts': 'Host',
            'switches': 'Switch',
            'controllers': 'Controller'
        }
        for name in titles:
            nodes = getattr( net, name )
            frame, consoles = self.createConsoles(
                cframe, nodes, width, titles[ name ] )
            self.consoles[ name ] = Object( frame=frame, consoles=consoles )
        self.selected = None
        self.select( 'hosts' )
        self.cframe.pack( expand=True, fill='both' )
        cleanUpScreens()
        # Close window gracefully
        Wm.wm_protocol( self.top, name='WM_DELETE_WINDOW', func=self.quit )
        self.pack( expand=True, fill='both' )
        
        

    
    def setOutputHook( self, fn=None, consoles=None ):
        "Register fn as output hook [on specific consoles.]"
        if consoles is None:
            consoles = self.consoles[ 'hosts' ].consoles
        for console in consoles:
            console.outputHook = fn

    def createConsoles( self, parent, nodes, width, title ):
        "Create a grid of consoles in a frame."
        f = Frame( parent )
        # Create consoles
        consoles = []
        index = 0
        for node in nodes:
            console = Console( f, self.net, node, title=title )
            consoles.append( console )
            row = index / width
            column = index % width
            console.grid( row=row, column=column, sticky='nsew' )
            index += 1
            f.rowconfigure( row, weight=1 )
            f.columnconfigure( column, weight=1 )
        return f, consoles

    def select( self, groupName ):
        "Select a group of consoles to display."
        if self.selected is not None:
            self.selected.frame.pack_forget()
        self.selected = self.consoles[ groupName ]
        self.selected.frame.pack( expand=True, fill='both' )

    def createMenuBar( self ):
        "Create and return a menu (really button) bar."
        f = Frame( self )
        buttons = [
            # ( 'Hosts', lambda: self.select( 'hosts' ) ),
            # ( 'Switches', lambda: self.select( 'switches' ) ),
            # ( 'Controllers', lambda: self.select( 'controllers' ) ),
            # # ( 'Graph', lambda: self.select( 'graph' ) ),
            ( 'DisconnectLink', self.disconnectLink ),
            ( 'ConnectAll', self.connectAllDisconnectedLinkd ),
            ( 'Interrupt', self.stop ),
            ( 'Clear', self.clear ),
            ( 'Quit', self.quit )
        ]
        for name, cmd in buttons:
            b = Button( f, text=name, command=cmd, **self.menuStyle )
            b.pack( side='left' )
        f.pack( padx=4, pady=4, fill='x' )
        return f

    def clear( self ):
        "Clear selection."
        for console in self.selected.consoles:
            console.clear()
            
        

    def waiting( self, consoles=None ):
        "Are any of our hosts waiting for output?"
        if consoles is None:
            consoles = self.consoles[ 'hosts' ].consoles
        for console in consoles:
            if console.waiting():
                return True
        return False

    def disconnectLink( self ):
        
        index = randint(0, len(self.net.links)-1)
        link = self.net.links[index]
        peerA = str(link).split('<->')[0].split('-')[0]
        peerB = str(link).split('<->')[1].split('-')[0]
        disLink = (peerA, peerB)
        self.net.configLinkStatus(peerA, peerB, 'down')
        self.disconnectedLinks.append(disLink)
        print peerA+'-'+peerB+' disconnected'
        if len(self.disconnectedLinks) > 10:
            conLink = self.disconnectedLinks.pop(0)
            self.net.configLinkStatus(conLink[0], conLink[1], 'up')
            print conLink[0]+'-'+conLink[1]+' connected'


     
    def connectAllDisconnectedLinkd( self ):
        while len(self.disconnectedLinks):
            conLink = self.disconnectedLinks.pop(0)
            self.net.configLinkStatus(conLink[0], conLink[1], 'up')
            print conLink[0]+'-'+conLink[1]+' connected'

        self.net.configLinkStatus('s3','h4','up')
        

    def stop( self, wait=True ):
        "Interrupt all hosts."
        consoles = self.consoles[ 'hosts' ].consoles
        for console in consoles:
            console.handleInt()
        if wait:
            for console in consoles:
                console.waitOutput()
        self.setOutputHook( None )
        # Shut down any iperfs that might still be running
        quietRun( 'killall -9 iperf' )

    def quit( self ):
        "Stop everything and quit."
        self.stop( wait=False)
        Frame.quit( self )


# Make it easier to construct and assign objects

def assign( obj, **kwargs ):
    "Set a bunch of fields in an object."
    obj.__dict__.update( kwargs )

class Object( object ):
    "Generic object you can stuff junk into."
    def __init__( self, **kwargs ):
        assign( self, **kwargs )


if __name__ == '__main__':
    setLogLevel( 'info' )
    network = TreeNet( depth=2, fanout=8 )
    network.start()
    app = ConsoleApp( network, width=4 )
    app.mainloop()
    network.stop()


