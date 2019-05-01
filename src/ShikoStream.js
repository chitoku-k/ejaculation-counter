const PQueue = require("p-queue");
const WebSocket = require("ws");
const { fromEvent, from, Observable } = require("rxjs");
const { filter, map, mergeMap, toArray } = require("rxjs/operators");

const RECONNECT_NONE = 0;
const RECONNECT_INITIAL = 5;
const RECONNECT_MAX = 320;

exports.ShikoStream = class ShikoStream {
    constructor(service, actions) {
        this.service = service;
        this.actions = actions;
        this.reconnect = 0;
        this.queue = new PQueue({
            concurrency: 1,
        });
    }

    create() {
        this.stream = Observable.create(observer => {
            const socket = new WebSocket(
                `${process.env.MASTODON_WSS_URL}streaming?access_token=${process.env.MASTODON_ACCESS_TOKEN}&stream=user`,
            );

            fromEvent(socket, "open").subscribe(() => {
                console.log("ShikoStream has established.");

                // Resets retry interval value when the connection is established
                this.reconnect = 0;
            });

            fromEvent(socket, "close").subscribe(() => {
                // Interval value will be doubled on every retry to a maximum of RECONNECT_MAX ms
                this.reconnect = Math.min(
                    Math.max(
                        this.reconnect * 2,
                        RECONNECT_INITIAL,
                    ),
                    RECONNECT_MAX,
                );

                console.warn(`ShikoStream was closed. Reconnecting in ${this.reconnect} s...`);
                setTimeout(() => this.create(), this.reconnect * 1000);
            });

            fromEvent(socket, "message").subscribe((...args) => observer.next(...args));
            fromEvent(socket, "error").subscribe((...args) => observer.error(...args));
        });

        this.stream.pipe(
            map(x => JSON.parse(x.data)),
            filter(x => x.event === "update"),
            map(x => JSON.parse(x.payload)),
            filter(x => !x.application || x.application.name !== process.env.MASTODON_APP),
            map(x => this.service.decodeToot(x)),
            mergeMap(toot => from(this.actions).pipe(
                map(action => ({
                    match: action.regex.exec(toot.content),
                    emoji: action.emoji && action.emoji.findIndex(x => toot.emojis.some(e => x === e.shortcode)),
                    action,
                    toot,
                })),
                filter(({ match, emoji }) => match || emoji >= 0),
                toArray(),
                map(x => x.sort((a, b) => a.emoji >= 0 || b.emoji >= 0 ? 1 : a.match.index - b.match.index)),
            )),
            mergeMap(x => x, (outer, inner) => inner),
        ).subscribe(
            ({ action, toot }) => this.queue.add(() => action.invoke(toot)),
            error => console.error(error.message),
        );
    }
}
