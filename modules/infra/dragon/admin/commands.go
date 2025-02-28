package admin

import (
	"dragon/common"
	"dragon/types"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"
	"github.com/jackc/pgtype"
	log "github.com/sirupsen/logrus"
)

const good = 0x00ff00
const bad = 0xe74c3c

var (
	commands         = make(map[string]types.AdminOp)
	commandNameCache = make(map[string]string)
)

// Admin OP Getter
func cmdInit() {
	// Requeue

	commands["REQUEUE"] = types.AdminOp{
		InternalName:   "requeue",
		Recursive:      false,
		Cooldown:       types.CooldownRequeue,
		Description:    "Requeue a bot",
		MinimumPerm:    3,
		ReasonNeeded:   true,
		Event:          types.EventBotRequeue,
		SlashSupported: true,
		SlashOptions: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "Reason for requeuing the bot",
				Required:    true,
			},
		},
		Server: common.TestServer,
		Handler: func(context types.AdminContext) string {
			if context.BotState != types.BotStateDenied {
				return "This bot cannot be requeued as it is not currently denied"
			}

			embed := discordgo.MessageEmbed{
				URL:         "https://fateslist.xyz/bot/" + context.Bot.ID,
				Title:       "Bot Requeued",
				Description: context.Bot.Mention() + " has been requeued (removed from the deny list)!",
				Color:       good,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Reason",
						Value: *context.Reason,
					},
				},
			}
			context.Discord.ChannelMessageSendComplex(common.SiteLogs, &discordgo.MessageSend{
				Embed: &embed,
			})
			_, err := context.Postgres.Exec(context.Context, "UPDATE bots SET state = $1 WHERE bot_id = $2", types.BotStatePending.Int(), context.Bot.ID)
			if err != nil {
				log.Error(err)
			}

			return ""
		},
	}
	// Claim
	commands["CLAIM"] = types.AdminOp{
		InternalName:   "claim",
		Recursive:      false,
		Cooldown:       types.CooldownNone,
		Description:    "Claim a bot",
		MinimumPerm:    2,
		ReasonNeeded:   false,
		Event:          types.EventBotClaim,
		SlashSupported: true,
		SlashOptions:   []*discordgo.ApplicationCommandOption{},
		Server:         common.TestServer,
		Handler: func(context types.AdminContext) string {
			if context.BotState != types.BotStatePending {
				return "This bot cannot be claimed as it is not currently pending review or it is already under review"
			}

			embed := discordgo.MessageEmbed{
				URL:         "https://fateslist.xyz/bot/" + context.Bot.ID,
				Title:       "Bot Under Review",
				Description: context.Bot.Mention() + " is now under review by " + context.User.Mention() + " and should be approved or denied soon!",
				Color:       good,
			}
			_, err := context.Discord.ChannelMessageSendComplex(common.SiteLogs, &discordgo.MessageSend{
				Content: "<@" + context.Owner + ">",
				Embed:   &embed,
			})

			if err != nil {
				log.Warn(err)
			}

			_, err = context.Postgres.Exec(context.Context, "UPDATE bots SET state = $1 WHERE bot_id = $2", types.BotStateUnderReview.Int(), context.Bot.ID)
			if err != nil {
				log.Error(err)
			}

			return ""
		},
	}

	// Unclaim
	commands["UNCLAIM"] = types.AdminOp{
		InternalName:   "unclaim",
		Recursive:      false,
		Cooldown:       types.CooldownNone,
		Description:    "Unclaim a bot",
		MinimumPerm:    2,
		ReasonNeeded:   false,
		Event:          types.EventBotUnclaim,
		SlashSupported: true,
		SlashOptions:   []*discordgo.ApplicationCommandOption{},
		Server:         common.TestServer,
		Handler: func(context types.AdminContext) string {
			if context.BotState != types.BotStateUnderReview {
				return "This bot cannot be unclaimed as it is not currently under review"
			}

			embed := discordgo.MessageEmbed{
				URL:         "https://fateslist.xyz/bot/" + context.Bot.ID,
				Title:       "Bot Unclaimed",
				Description: context.Bot.Mention() + " has been unclaimed by " + context.User.Mention() + ". It is no longer under review right now but it should be approved or denied when another reviewer comes in! Don't worry, this is completely normal!",
				Color:       good,
			}
			_, err := context.Discord.ChannelMessageSendComplex(common.SiteLogs, &discordgo.MessageSend{
				Embed: &embed,
			})

			if err != nil {
				log.Warn(err)
			}

			_, err = context.Postgres.Exec(context.Context, "UPDATE bots SET state = $1 WHERE bot_id = $2", types.BotStatePending.Int(), context.Bot.ID)
			if err != nil {
				log.Error(err)
			}

			return ""
		},
	}

	// Ban
	commands["BAN"] = types.AdminOp{
		InternalName:   "ban",
		Recursive:      false,
		Cooldown:       types.CooldownBan,
		Description:    "Bans a bot",
		MinimumPerm:    3,
		ReasonNeeded:   true,
		Event:          types.EventBotBan,
		SlashSupported: true,
		SlashOptions: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "Reason to ban the bot",
				Required:    true,
			},
		},
		Server: common.StaffServer,
		Handler: func(context types.AdminContext) string {
			embed := discordgo.MessageEmbed{
				URL:         "https://fateslist.xyz/bot/" + context.Bot.ID,
				Title:       "Bot Banned",
				Description: context.Bot.Mention() + " has been banned by " + context.User.Mention() + ".",
				Color:       bad,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Reason",
						Value: *context.Reason,
					},
				},
			}
			_, err := context.Discord.ChannelMessageSendComplex(common.SiteLogs, &discordgo.MessageSend{
				Content: "<@" + context.Owner + ">",
				Embed:   &embed,
			})

			if err != nil {
				log.Warn(err)
			}
			_, err = context.Postgres.Exec(context.Context, "UPDATE bots SET state = $1, verifier = $2 WHERE bot_id = $3", types.BotStateBanned.Int(), context.User.ID, context.Bot.ID)

			if err != nil {
				log.Error(err)
				return "PostgreSQL error: " + err.Error()
			}

			err = context.Discord.GuildBanCreateWithReason(common.MainServer, context.Bot.ID, *context.Reason, 0)

			if err != nil {
				return "OK. Bot was banned successfully but it could not be kicked due to reason: " + err.Error()
			}
			return ""
		},
	}

	// Unban
	commands["UNBAN"] = types.AdminOp{
		InternalName:   "unban",
		Recursive:      false,
		Cooldown:       types.CooldownBan,
		Description:    "Unbans a bot",
		MinimumPerm:    3,
		ReasonNeeded:   true,
		Event:          types.EventBotUnban,
		SlashSupported: true,
		SlashOptions: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "Reason to unban the bot",
				Required:    true,
			},
		},
		Server: common.StaffServer,
		Handler: func(context types.AdminContext) string {
			embed := discordgo.MessageEmbed{
				URL:         "https://fateslist.xyz/bot/" + context.Bot.ID,
				Title:       "Bot Unbanned",
				Description: context.Bot.Mention() + " has been unbanned by " + context.User.Mention() + ".",
				Color:       good,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Extra Info/Reason",
						Value: *context.Reason,
					},
				},
			}
			_, err := context.Discord.ChannelMessageSendComplex(common.SiteLogs, &discordgo.MessageSend{
				Content: "<@" + context.Owner + ">",
				Embed:   &embed,
			})

			if err != nil {
				log.Warn(err)
			}

			_, err = context.Postgres.Exec(context.Context, "UPDATE bots SET state = $1, verifier = $2 WHERE bot_id = $3", types.BotStateApproved.Int(), context.User.ID, context.Bot.ID)

			if err != nil {
				log.Error(err)
				return "Got an error when trying to update the database. Please report this: " + err.Error()
			}

			context.Discord.GuildBanDelete(common.MainServer, context.Bot.ID)
			return ""
		},
	}

	// Certify
	commands["CERTIFY"] = types.AdminOp{
		InternalName:   "certify",
		Recursive:      false,
		Cooldown:       types.CooldownNone,
		Description:    "Certifies a bot",
		MinimumPerm:    5,
		ReasonNeeded:   false,
		Event:          types.EventBotCertify,
		SlashSupported: true,
		SlashOptions:   []*discordgo.ApplicationCommandOption{},
		Server:         common.StaffServer,
		Handler: func(context types.AdminContext) string {
			var errors string = "OK. "
			if context.BotState != types.BotStateApproved {
				return "This bot cannot be certified as it is not approved or is already certified"
			}

			embed := discordgo.MessageEmbed{
				URL:         "https://fateslist.xyz/bot/" + context.Bot.ID,
				Title:       "Bot Certified",
				Description: context.Bot.Mention() + " has been certified by " + context.User.Mention() + ". Congratulations on your accompishment :heart:",
				Color:       good,
			}
			_, err := context.Discord.ChannelMessageSendComplex(common.SiteLogs, &discordgo.MessageSend{
				Content: "<@" + context.Owner + ">",
				Embed:   &embed,
			})

			if err != nil {
				log.Warn(err)
			}

			_, err = context.Postgres.Exec(context.Context, "UPDATE bots SET state = $1 WHERE bot_id = $3", types.BotStateCertified.Int(), context.Bot.ID)

			if err != nil {
				log.Error(err)
			}

			err = context.Discord.GuildMemberRoleAdd(common.MainServer, context.Bot.ID, common.CertifiedBotRole)

			if err != nil {
				errors += err.Error() + "\n"
			}

			// Give certified bot role
			owners, err := context.Postgres.Query(context.Context, "SELECT owner FROM bot_owner WHERE bot_id = $1", context.Bot.ID)

			if err != nil {
				errors += "Bot was certified, but I could not find bot owners because: " + err.Error() + "\n"
				return "OK. Bot was certified, but I could not add certified dev role to bot owners because: \n" + errors
			}

			defer owners.Close()

			var i int

			for owners.Next() {
				i += 1
				var owner pgtype.Int8
				err = owners.Scan(&owner)
				if err != nil {
					errors += "Got error: " + err.Error() + " in iteration " + strconv.Itoa(i) + " (user id is unknown)\n"
					continue
				}

				if owner.Status != pgtype.Present {
					errors += "Got error: Owner is NULL in iteration" + strconv.Itoa(i) + "\n"
					continue
				}

				err = context.Discord.GuildMemberRoleAdd(common.MainServer, strconv.FormatInt(owner.Int, 10), common.CertifiedDevRole)
				if err != nil {
					errors += "Got error: " + err.Error() + " in iteration " + strconv.Itoa(i) + "and user id (" + strconv.FormatInt(owner.Int, 10) + ")\n"
					continue
				}
			}
			if errors != "OK. " {
				return "OK. Bot was certified, but I could not add certified bot role to bot owners because: \n" + errors
			}

			return ""
		},
	}

	// Uncertify
	commands["UNCERTIFY"] = types.AdminOp{
		InternalName:   "uncertify",
		Recursive:      false,
		Cooldown:       types.CooldownNone,
		Description:    "Uncertifies a bot",
		MinimumPerm:    5,
		ReasonNeeded:   true,
		Event:          types.EventBotUncertify,
		SlashSupported: true,
		SlashOptions: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "Reason for uncertifying this bot",
				Required:    true,
			},
		},
		Server: common.StaffServer,
		Handler: func(context types.AdminContext) string {
			var errors string = "OK. "
			if context.BotState != types.BotStateCertified {
				return "This bot cannot be uncertified as it is not currently certified"
			}

			embed := discordgo.MessageEmbed{
				URL:         "https://fateslist.xyz/bot/" + context.Bot.ID,
				Title:       "Bot Uncertified",
				Description: context.Bot.Mention() + " has been uncertified by " + context.User.Mention() + ".",
				Color:       bad,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Reason",
						Value: *context.Reason,
					},
				},
			}
			_, err := context.Discord.ChannelMessageSendComplex(common.SiteLogs, &discordgo.MessageSend{
				Content: "<@" + context.Owner + ">",
				Embed:   &embed,
			})

			if err != nil {
				log.Warn(err)
			}

			_, err = context.Postgres.Exec(context.Context, "UPDATE bots SET state = $1 WHERE bot_id = $3", types.BotStateApproved.Int(), context.Bot.ID)

			if err != nil {
				log.Error(err)
			}

			err = context.Discord.GuildMemberRoleRemove(common.MainServer, context.Bot.ID, common.CertifiedBotRole)

			if err != nil {
				errors += "Failed to remove certified bot role because: " + err.Error() + "\n"
			}
			return ""
		},
	}

	// Approve
	commands["APPROVE"] = types.AdminOp{
		InternalName:      "approve",
		Recursive:         false,
		Cooldown:          types.CooldownNone,
		Description:       "Approves a bot",
		MinimumPerm:       2,
		ReasonNeeded:      true,
		Event:             types.EventBotApprove,
		SlashSupported:    true,
		SlashContextField: "guild_count",
		SlashOptions: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "Feedback about the bot and/or a welcome message",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "guild_count",
				Description: "Guild count of the bot currently",
				Required:    true,
			},
		},
		Server: common.TestServer,
		Handler: func(context types.AdminContext) string {
			if context.BotState != types.BotStateUnderReview {
				return "This bot cannot be approved as it is not currently under review. Did you claim it first?"
			} else if context.ExtraContext == nil {
				return "Bots approximate guild count must be set when approving"
			}

			guild_count, err := strconv.Atoi(*context.ExtraContext)

			if err != nil {
				return "Could not parse guild count: " + err.Error()
			}

			var errors string = "OK. "

			embed := discordgo.MessageEmbed{
				URL:         "https://fateslist.xyz/bot/" + context.Bot.ID,
				Title:       "Bot Approved",
				Description: context.Bot.Mention() + " has been approved by " + context.User.Mention() + ". Congratulations on your accompishment :heart:",
				Color:       good,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Feedback",
						Value: *context.Reason,
					},
				},
			}
			_, err = context.Discord.ChannelMessageSendComplex(common.SiteLogs, &discordgo.MessageSend{
				Content: "<@" + context.Owner + ">",
				Embed:   &embed,
			})

			if err != nil {
				log.Warn(err)
			}

			_, err = context.Postgres.Exec(context.Context, "UPDATE bots SET state = $1, verifier = $2, guild_count = $3 WHERE bot_id = $4", types.BotStateApproved.Int(), context.User.ID, guild_count, context.Bot.ID)

			if err != nil {
				log.Error(err)
				return "Got an error when trying to update the database. Please report this: " + err.Error()
			}

			// Give bot bot role
			owners, err := context.Postgres.Query(context.Context, "SELECT owner FROM bot_owner WHERE bot_id = $1", context.Bot.ID)

			if err != nil {
				errors += "Bot was approved, but I could not find bot owners because: " + err.Error() + "\n"
				return "OK. Bot was approved, but I could not add bot dev role to bot owners because: \n" + errors
			}

			defer owners.Close()

			var i int

			for owners.Next() {
				i += 1
				var owner pgtype.Int8
				err = owners.Scan(&owner)
				if err != nil {
					errors += "Got error: " + err.Error() + " in iteration " + strconv.Itoa(i) + " (user id is unknown)\n"
					continue
				}

				if owner.Status != pgtype.Present {
					errors += "Got error: Owner is NULL in iteration" + strconv.Itoa(i) + "\n"
					continue
				}

				err = context.Discord.GuildMemberRoleAdd(common.MainServer, strconv.FormatInt(owner.Int, 10), common.BotDevRole)
				if err != nil {
					errors += "Got error: " + err.Error() + " in iteration " + strconv.Itoa(i) + "and user id (" + strconv.FormatInt(owner.Int, 10) + ")\n"
					continue
				}
			}

			if errors != "OK. " {
				return "OK. Bot was approved, but I could not add bot dev role to bot owners because: \n" + errors
			}

			return ""
		},
	}

	// Denies a bot
	commands["DENY"] = types.AdminOp{
		InternalName:   "deny",
		Recursive:      false,
		Cooldown:       types.CooldownNone,
		Description:    "Denies a bot",
		MinimumPerm:    2,
		ReasonNeeded:   true,
		Event:          types.EventBotDeny,
		SlashSupported: true,
		SlashOptions: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "Reason for denying the bot",
				Required:    true,
			},
		},
		Server: common.TestServer,
		Handler: func(context types.AdminContext) string {
			if context.BotState != types.BotStateUnderReview {
				return "This bot cannot be denied as it is not currently under review. Did you claim it first?"
			}

			embed := discordgo.MessageEmbed{
				URL:         "https://fateslist.xyz/bot/" + context.Bot.ID,
				Title:       "Bot Denied",
				Description: context.Bot.Mention() + " has been denied by " + context.User.Mention() + ".",
				Color:       bad,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Reason",
						Value: *context.Reason,
					},
				},
			}
			_, err := context.Discord.ChannelMessageSendComplex(common.SiteLogs, &discordgo.MessageSend{
				Content: "<@" + context.Owner + ">",
				Embed:   &embed,
			})

			if err != nil {
				log.Warn(err)
			}

			_, err = context.Postgres.Exec(context.Context, "UPDATE bots SET state = $1, verifier = $2 WHERE bot_id = $3", types.BotStateDenied.Int(), context.User.ID, context.Bot.ID)

			if err != nil {
				log.Error(err)
			}

			return ""
		},
	}

	// Unverifies a bot
	commands["UNVERIFY"] = types.AdminOp{
		InternalName:   "unverify",
		Recursive:      false,
		Cooldown:       types.CooldownNone,
		Description:    "Unverifies a bot",
		MinimumPerm:    2,
		ReasonNeeded:   true,
		Event:          types.EventBotUnverify,
		SlashSupported: true,
		SlashOptions: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "Reason for unverifying the bot",
				Required:    true,
			},
		},
		Server: common.StaffServer,
		Handler: func(context types.AdminContext) string {
			if context.BotState != types.BotStateApproved {
				return "This bot cannot be unverified as it is not currently approved or is certified."
			}

			embed := discordgo.MessageEmbed{
				URL:         "https://fateslist.xyz/bot/" + context.Bot.ID,
				Title:       "Bot Unverified",
				Description: context.Bot.Mention() + " has been unverified by " + context.User.Mention() + ".",
				Color:       bad,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Reason",
						Value: *context.Reason,
					},
				},
			}
			_, err := context.Discord.ChannelMessageSendComplex(common.SiteLogs, &discordgo.MessageSend{
				Content: "<@" + context.Owner + ">",
				Embed:   &embed,
			})

			if err != nil {
				log.Warn(err)
			}

			_, err = context.Postgres.Exec(context.Context, "UPDATE bots SET state = $1, verifier = $2 WHERE bot_id = $3", types.BotStateUnderReview.Int(), context.User.ID, context.Bot.ID)

			if err != nil {
				log.Error(err)
			}

			return ""
		},
	}

	commands["STAFFLOCK"] = types.AdminOp{
		InternalName:   "stafflock",
		Recursive:      false,
		Cooldown:       types.CooldownNone,
		Description:    "Staff locks a bot",
		MinimumPerm:    2,
		ReasonNeeded:   false,
		Event:          types.EventStaffLock,
		SlashSupported: true,
		SlashOptions:   []*discordgo.ApplicationCommandOption{},
		Server:         common.StaffServer,
		Handler: func(context types.AdminContext) string {
			countKey := "fl_staff_access-" + context.User.ID + ":count"
			accessKey := "fl_staff_access-" + context.User.ID + ":" + context.Bot.ID

			botLocked := context.Redis.Exists(context.Context, accessKey).Val()
			if botLocked == 0 {
				return "You have not unlocked this bot yet!"
			}
			pipeline := context.Redis.Pipeline()
			pipeline.Decr(context.Context, countKey)
			pipeline.Del(context.Context, accessKey)
			_, err := pipeline.Exec(context.Context)
			if err != nil {
				log.Warn(err)
				return "Something happened! " + err.Error()
			}

			embed := discordgo.MessageEmbed{
				URL:         "https://fateslist.xyz/bot/" + context.Bot.ID,
				Title:       "Staff Lock",
				Description: context.Bot.Mention() + " has been locked by " + context.User.Mention() + ". This is perfectly normal and is a safety measure against hacking and exploits",
				Color:       good,
			}

			_, err = context.Discord.ChannelMessageSendComplex(common.SiteLogs, &discordgo.MessageSend{
				Content: "<@" + context.Owner + ">",
				Embed:   &embed,
			})

			if err != nil {
				log.Warn(err)
			}

			return "Thank you for relocking this bot and keeping Fates List safe"
		},
	}

	commands["STAFFUNLOCK"] = types.AdminOp{
		InternalName:   "staffunlock",
		Recursive:      false,
		Cooldown:       types.CooldownNone,
		Description:    "Staff unlocks a bot",
		MinimumPerm:    2,
		ReasonNeeded:   true,
		Event:          types.EventStaffUnlock,
		SlashSupported: true,
		SlashOptions: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "Reason you need to staff unlock the bot, is publicly visible",
				Required:    true,
			},
		},
		Server: common.StaffServer,
		Handler: func(context types.AdminContext) string {
			countKey := "fl_staff_access-" + context.User.ID + ":count"
			accessKey := "fl_staff_access-" + context.User.ID + ":" + context.Bot.ID

			botLocked := context.Redis.Exists(context.Context, accessKey).Val()
			if botLocked != 0 {
				return "You have already locked this bot."
			}
			botLockedCount, err := context.Redis.Get(context.Context, countKey).Int()

			if err == nil && botLockedCount > 0 {
				return "You may only have one bot unlocked at any given time!"
			}

			pipeline := context.Redis.Pipeline()

			pipeline.Incr(context.Context, countKey)
			pipeline.Expire(context.Context, countKey, 60*time.Minute)
			pipeline.Set(context.Context, accessKey, "0", 30*time.Minute)
			_, err = pipeline.Exec(context.Context)
			if err != nil {
				log.Warn(err)
				return "Something happened! " + err.Error()
			}

			embed := discordgo.MessageEmbed{
				URL:         "https://fateslist.xyz/bot/" + context.Bot.ID,
				Title:       "Staff Unlock",
				Description: context.Bot.Mention() + " has been unlocked by " + context.User.Mention() + ". This is perfectly normal This is normal but if it happens too much, open a ticket or otherwise contact any online or offline staff immediately",
				Color:       good,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Reason",
						Value: *context.Reason,
					},
				},
			}

			_, err = context.Discord.ChannelMessageSendComplex(common.SiteLogs, &discordgo.MessageSend{
				Content: "<@" + context.Owner + ">",
				Embed:   &embed,
			})

			if err != nil {
				log.Warn(err)
			}

			return "OK. Be **absolutely** *sure* to relock the bot as soon as possible"
		},
	}

	// Load command name cache to map internal name to the command
	for cmdName, v := range commands {
		commandNameCache[v.InternalName] = cmdName
	}
}

func GetCommandSpew() string {
	return spew.Sdump("Admin commands loaded: ", commands)
}
