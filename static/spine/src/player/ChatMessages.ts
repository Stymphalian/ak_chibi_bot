// TODO: Make these configurable
const SHOW_DURATION_SECS = 5;
const MAX_CHARS_PER_LINE = 30;
const MAX_LINES_PER_BLOCK = 3;

export class MessageBlock {
    public messages: string[] = [];
    constructor(messages: string[]) {
        this.messages = messages;
    }
}

export class ChatMessageQueue {
    private messages: MessageBlock[] = [];
    private currentMessage: MessageBlock|null = null;    
    private showTimeRemaining: number = 0;
    private showing: boolean = false;

    public AddMessage(message: string) {
        let splitMessages = message.split(/\s+/);
        splitMessages = splitMessages.flatMap((line) => {
            if (line.length > MAX_CHARS_PER_LINE) {
                return line.match(new RegExp(`.{1,${MAX_CHARS_PER_LINE}}`, 'g'));
            } else {
                return line;
            }
        });

        let newLines = [];
        let currentLine = "";
        for (let msg of splitMessages) {
            if (currentLine.length + msg.length + 1 > MAX_CHARS_PER_LINE) {
                newLines.push(currentLine);
                currentLine = msg;
            } else {
                currentLine += " " + msg;
            }
        }
        if (currentLine != "") {
            newLines.push(currentLine);
        }

        for (let i = 0; i < newLines.length; i += MAX_LINES_PER_BLOCK) {
            this.messages.push(new MessageBlock(
                newLines.slice(i, i + MAX_LINES_PER_BLOCK)
            ));      
        }                
    }

    public HasMessages(): boolean {
        return this.messages.length > 0
    }

    public GetCurrentMessageBlock(): MessageBlock|null {
        return this.currentMessage;
    }

    public Update(deltaSecs: number) {
        if (this.currentMessage == null) {
            if (this.HasMessages()) {
                this.currentMessage = this.messages.shift();
                this.showTimeRemaining = SHOW_DURATION_SECS;
                this.showing = true;
            }
            return;
        }
        if (this.showing) {
            this.showTimeRemaining -= deltaSecs;
            if (this.showTimeRemaining <= 0) {
                if (this.messages.length > 0) {
                    this.currentMessage = this.messages.shift();
                    this.showTimeRemaining = SHOW_DURATION_SECS;
                } else {
                    this.showing = false;
                    this.currentMessage = null;
                }
            }
        }
    }
}