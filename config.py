from config_secrets import *
bot_logs = 836326348326043648 # #bot-logs in support server
server_logs = 837048691965034496 # Server logs in support server
appeals_channel = 836326351387623454 # #resubmissions-and-appeals in support server
site_errors_channel = 836326323281592330 # Where to log site errors
bots_role = 836326315946672158 # BOTS role in support server
staff_ping_add_role=836326316188893275
bot_dev_role = 836326314344185876 # BOT Developer in support server
certified_dev_role = 836326313283026995 # Certified Developer in support server
main_server = 789934742128558080 # Main server
staff_server = 816130947274899487 # The staff server
staff_ag = 845931695387181066 # Access Granted role in staff server
test_botsrole = 845940351339987004 # Bots role on test server
test_staffrole = 846028433985503232 # Staff role on test server
test_server = 845928429357367316 # Test Server
owner = 563808552288780322

# Messages
approve_feedback = "There was no feedback given for this bot. It was likely a good bot, but you can ask any staff member about feedback if you wish"
deny_feedback = "There was no reason specified. DM/Ping Mod Mail to learn why"

# Banned in vanity

reserved_vanity = ["bot", "docs", "redoc", "doc", "profile", "server", "bots", "servers", "search", "invite", "discord", "login", "logout", "register", "admin", "console", "fuck", "support"]

staff_roles = {
    "user": {
        "id": 00000000000000000,
        "staff_id": 0000000000000000,
        "perm": 0
    },
    "auditor": {
        "id": 000000000000000000,
        "staff_id": 0000000000000000,
        "perm": 1
    }, # Unused
    "bot_reviewer": {
        "id": 836326311147864104,
        "staff_id": 845931373520748545,
        "perm": 2
    },
    "mod": {
        "id": 836326309528600627,
        "staff_id": 845931545076432931,
        "perm": 3
    },
    "admin": {
        "id": 836326305666039848,
        "staff_id": 845930903883874336,
        "perm": 4,
    },
    "head_admin": {
        "id": 836349482340843572,
        "staff_id": 845930541018513428,
        "perm": 5
    },
    "owner": {
        "id": 836326299223195738,
        "staff_id": 830540676952227863,
        "perm": 6,
    }
}
# TODO: Add Timed Badges
special_badges = {
    "STAFF": {
        "name": "Staff",
        "description": "This is a Fates List Staff Member",
        "image": "/static/assets/img/staff.png",
        "staff": True, # Is this badge only for staff?
        "certified": False # Certified
    },
    "CERTIFIED": {
        "name": "Certified Bot Dev.",
        "description": "This is a certified bot developer",
        "image": "/static/assets/img/certified.png",
        "staff": False, # Is this badge only for staff?
        "certified": True # Certified
    }
}

features = {
    "custom_prefix": {
        "type": "positive",
        "description": "A bot with Custom Prefix supports changing of the bots prefix and is considered positive by Fates List"
    },
    "open_source": {
        "type": "positive",
        "description": "These bots are open source meaning they can easily be audited and/or potentially self hosted."
    }
} 

site = "fateslist.xyz" # Replace this with your domain

server_bot_invite = "https://discord.com/api/oauth2/authorize?client_id=811073947382579200&permissions=67649&scope=bot" # Ensure that it uses 67649 for perms

# This value below dont need to be changed
site_url = "https://" + site
support_url = "https://discord.gg/cMAnfu8AJB"
#TAGS = {"music": "fa-solid:music", "moderation": "fa-solid:hammer", "economy": "fa-solid:coins", "fun": "fa-solid:heart", "anime": "fa-solid:camera", "games": "fa-solid:smile", "web_dashboard": "fa-solid:chart-bar", "logging": "fa-solid:chart-line", "game_stats": "fa-solid:chart-pie", "leveling": "fa-solid:arrow-up", "roleplay": "fa-solid:cat", "utility": "fa-solid:desktop", "social": "fa-solid:users", "meme": "cib:happycow", "reddit": "akar-icons:reddit-fill", "pokemon": "mdi-pokemon-go"}
SERVER_TAGS = {"music": "fa-solid:music", "moderation": "fa-solid:hammer", "economy": "fa-solid:coins", "fun": "fa-solid:heart", "anime": "fa-solid:camera", "games": "fa-solid:smile", "web_dashboard": "fa-solid:chart-bar", "logging": "fa-solid:chart-line", "game_stats": "fa-solid:chart-pie", "leveling": "fa-solid:arrow-up", "roleplay": "fa-solid:cat", "utility": "fa-solid:desktop", "social": "fa-solid:users"}


pg_user = "postgres" # Postgres Database Username

bans_data = {
    "1": {
        "type": "global",
        "desc": "you cannot/will not be able to login or access the support server"
    },
    "2": {
        "type": "login",
        "desc": "you will not be able to login but should still be able to access the support server" 
    },
    "3": {
        "type": "profile edit",
        "desc": "you will not be able to edit your profile"
    },
    "4": {
        "type": "data deletion request",
        "desc": "Contact modmail to be unbanned"
    }
}

class OauthConfig:
    client_id = "798951566634778641"
    client_secret = oauth_client_secret
    redirect_uri = "https://" + site + "/auth/login/confirm"


