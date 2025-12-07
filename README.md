# instagrab

Download stuff from Instagram

Feature request or there's an issue the wiki doesn't cover? Join the Discord [here](https://discord.gg/yzNM7CPn4s)

## Installation

To build from source:

Clone repo
```bash
git clone https://github.com/horsaen/instagrab.git
```

Install
```bash
go install
```

## Usage

```bash
instagrab -username USR123 -mode story
```

-mode:
reels
posts
highlights
story

## Cookies/Auth
User cookies are required to interface with Instagram in any meaningful capacity

Found in your home folder @ `.instagrab/cookies`, you can input your corresponding cookies.

Line 1: X-CSRFToken
Line 2: Cookie