package tbc

import (
	"github.com/wowsims/tbc/sim/common/shared"
)

func RegisterAllOnUseCds() {

	//
	// shared.NewSimpleStatActive(13143) // Mark of the Dragon Lord - https://www.wowhead.com/tbc/spell=17252
	// shared.NewSimpleStatActive(13171) // Smokey's Lighter - https://www.wowhead.com/tbc/spell=17283
	// shared.NewSimpleStatActive(13213) // Smolderweb's Eye - https://www.wowhead.com/tbc/spell=17330
	// shared.NewSimpleStatActive(13515) // Ramstein's Lightning Bolts - https://www.wowhead.com/tbc/spell=17668
	// shared.NewSimpleStatActive(13937) // Headmaster's Charge - https://www.wowhead.com/tbc/spell=18264
	// shared.NewSimpleStatActive(14152) // Robe of the Archmage - https://www.wowhead.com/tbc/spell=18385
	// shared.NewSimpleStatActive(14153) // Robe of the Void - https://www.wowhead.com/tbc/spell=18386
	// shared.NewSimpleStatActive(17759) // Mark of Resolution - https://www.wowhead.com/tbc/spell=21956
	// shared.NewSimpleStatActive(18639) // Ultra-Flash Shadow Reflector - https://www.wowhead.com/tbc/spell=23132
	// shared.NewSimpleStatActive(19336) // Arcane Infused Gem - https://www.wowhead.com/tbc/spell=23721
	// shared.NewSimpleStatActive(19340) // Rune of Metamorphosis - https://www.wowhead.com/tbc/spell=23724
	// shared.NewSimpleStatActive(19341) // Lifegiving Gem - https://www.wowhead.com/tbc/spell=23725
	// shared.NewSimpleStatActive(19342) // Venomous Totem - https://www.wowhead.com/tbc/spell=23726
	// shared.NewSimpleStatActive(19930) // Mar'li's Eye - https://www.wowhead.com/tbc/spell=24268
	// shared.NewSimpleStatActive(19948) // Zandalarian Hero Badge - https://www.wowhead.com/tbc/spell=24574
	// shared.NewSimpleStatActive(19949) // Zandalarian Hero Medallion - https://www.wowhead.com/tbc/spell=24661
	// shared.NewSimpleStatActive(19950) // Zandalarian Hero Charm - https://www.wowhead.com/tbc/spell=24658
	// shared.NewSimpleStatActive(19951) // Gri'lek's Charm of Might - https://www.wowhead.com/tbc/spell=24571
	// shared.NewSimpleStatActive(19953) // Renataki's Charm of Beasts - https://www.wowhead.com/tbc/spell=24531
	// shared.NewSimpleStatActive(19954) // Renataki's Charm of Trickery - https://www.wowhead.com/tbc/spell=24532
	// shared.NewSimpleStatActive(19956) // Wushoolay's Charm of Spirits - https://www.wowhead.com/tbc/spell=24499
	// shared.NewSimpleStatActive(20071) // Talisman of Arathor - https://www.wowhead.com/tbc/spell=23991
	// shared.NewSimpleStatActive(20072) // Defiler's Talisman - https://www.wowhead.com/tbc/spell=23991
	// shared.NewSimpleStatActive(21181) // Grace of Earth - https://www.wowhead.com/tbc/spell=25892
	// shared.NewSimpleStatActive(21488) // Fetish of Chitinous Spikes - https://www.wowhead.com/tbc/spell=26168
	// shared.NewSimpleStatActive(21625) // Scarab Brooch - https://www.wowhead.com/tbc/spell=26467
	// shared.NewSimpleStatActive(21647) // Fetish of the Sand Reaver - https://www.wowhead.com/tbc/spell=26400
	// shared.NewSimpleStatActive(21685) // Petrified Scarab - https://www.wowhead.com/tbc/spell=26463
	// shared.NewSimpleStatActive(21891) // Shard of the Fallen Star - https://www.wowhead.com/tbc/spell=26789
	// shared.NewSimpleStatActive(23001) // Eye of Diminution - https://www.wowhead.com/tbc/spell=28862
	// shared.NewSimpleStatActive(23027) // Warmth of Forgiveness - https://www.wowhead.com/tbc/spell=28760
	// shared.NewSimpleStatActive(23040) // Glyph of Deflection - https://www.wowhead.com/tbc/spell=28773
	// shared.NewSimpleStatActive(23558) // The Burrower's Shell - https://www.wowhead.com/tbc/spell=29506
	// shared.NewSimpleStatActive(23564) // Twisting Nether Chain Shirt - https://www.wowhead.com/tbc/spell=34518
	// shared.NewSimpleStatActive(23565) // Embrace of the Twisting Nether - https://www.wowhead.com/tbc/spell=34518
	// shared.NewSimpleStatActive(23570) // Jom Gabbar - https://www.wowhead.com/tbc/spell=29602
	// shared.NewSimpleStatActive(23587) // Mirren's Drinking Hat - https://www.wowhead.com/tbc/spell=29830
	// shared.NewSimpleStatActive(23824) // Rocket Boots Xtreme - https://www.wowhead.com/tbc/spell=51582
	// shared.NewSimpleStatActive(23825) // Nigh Invulnerability Belt - https://www.wowhead.com/tbc/spell=30458
	// shared.NewSimpleStatActive(23835) // Gnomish Poultryizer - https://www.wowhead.com/tbc/spell=30507
	// shared.NewSimpleStatActive(23836) // Goblin Rocket Launcher - https://www.wowhead.com/tbc/spell=46567
	// shared.NewSimpleStatActive(24092) // Pendant of Frozen Flame - https://www.wowhead.com/tbc/spell=30997
	// shared.NewSimpleStatActive(24093) // Pendant of Thawing - https://www.wowhead.com/tbc/spell=30994
	// shared.NewSimpleStatActive(24095) // Pendant of Withering - https://www.wowhead.com/tbc/spell=30999
	// shared.NewSimpleStatActive(24097) // Pendant of Shadow's End - https://www.wowhead.com/tbc/spell=31000
	// shared.NewSimpleStatActive(24098) // Pendant of the Null Rune - https://www.wowhead.com/tbc/spell=31002
	// shared.NewSimpleStatActive(24106) // Thick Felsteel Necklace - https://www.wowhead.com/tbc/spell=31023
	// shared.NewSimpleStatActive(24110) // Living Ruby Pendant - https://www.wowhead.com/tbc/spell=31024
	// shared.NewSimpleStatActive(24116) // Eye of the Night - https://www.wowhead.com/tbc/spell=31033
	// shared.NewSimpleStatActive(24117) // Embrace of the Dawn - https://www.wowhead.com/tbc/spell=31026
	// shared.NewSimpleStatActive(24121) // Chain of the Twilight Owl - https://www.wowhead.com/tbc/spell=31035
	// shared.NewSimpleStatActive(24127) // Figurine - Talasite Owl - https://www.wowhead.com/tbc/spell=31045
	// shared.NewSimpleStatActive(24376) // Runed Fungalcap - https://www.wowhead.com/tbc/spell=31771
	// shared.NewSimpleStatActive(24390) // Auslese's Light Channeler - https://www.wowhead.com/tbc/spell=31794
	// shared.NewSimpleStatActive(24551) // Talisman of the Horde - https://www.wowhead.com/tbc/spell=32140
	// shared.NewSimpleStatActive(25786) // Hypnotist's Watch - https://www.wowhead.com/tbc/spell=32599
	// shared.NewSimpleStatActive(25827) // Muck-Covered Drape - https://www.wowhead.com/tbc/spell=32641
	// shared.NewSimpleStatActive(25829) // Talisman of the Alliance - https://www.wowhead.com/tbc/spell=33828
	// shared.NewSimpleStatActive(25996) // Emblem of Perseverance - https://www.wowhead.com/tbc/spell=32957
	// shared.NewSimpleStatActive(26055) // Oculus of the Hidden Eye - https://www.wowhead.com/tbc/spell=33012
	// shared.NewSimpleStatActive(27416) // Fetish of the Fallen - https://www.wowhead.com/tbc/spell=33014
	// shared.NewSimpleStatActive(27529) // Figurine of the Colossus - https://www.wowhead.com/tbc/spell=33089
	// shared.NewSimpleStatActive(27770) // Argussian Compass - https://www.wowhead.com/tbc/spell=39228
	// shared.NewSimpleStatActive(27900) // Jewel of Charismatic Mystique - https://www.wowhead.com/tbc/spell=33486
	// shared.NewSimpleStatActive(28042) // Regal Protectorate - https://www.wowhead.com/tbc/spell=33668
	// shared.NewSimpleStatActive(28590) // Ribbon of Sacrifice - https://www.wowhead.com/tbc/spell=38332
	// shared.NewSimpleStatActive(28727) // Pendant of the Violet Eye - https://www.wowhead.com/tbc/spell=29601
	// shared.NewSimpleStatActive(28767) // The Decapitator - https://www.wowhead.com/tbc/spell=37208
	// shared.NewSimpleStatActive(29181) // Timelapse Shard - https://www.wowhead.com/tbc/spell=35352
	// shared.NewSimpleStatActive(30841) // Lower City Prayerbook - https://www.wowhead.com/tbc/spell=37877
	// shared.NewSimpleStatActive(30847) // X-52 Rocket Helmet - https://www.wowhead.com/tbc/spell=37896
	// shared.NewSimpleStatActive(32501) // Shadowmoon Insignia - https://www.wowhead.com/tbc/spell=40464
	// shared.NewSimpleStatActive(32538) // Skywitch's Drape - https://www.wowhead.com/tbc/spell=12438
	// shared.NewSimpleStatActive(32539) // Skyguard's Drape - https://www.wowhead.com/tbc/spell=12438
	// shared.NewSimpleStatActive(33808) // The Horseman's Helm - https://www.wowhead.com/tbc/spell=43873
	// shared.NewSimpleStatActive(33820) // Weather-Beaten Fishing Hat - https://www.wowhead.com/tbc/spell=43699
	// shared.NewSimpleStatActive(34428) // Steely Naaru Sliver - https://www.wowhead.com/tbc/spell=45049
	// shared.NewSimpleStatActive(34429) // Shifting Naaru Sliver - https://www.wowhead.com/tbc/spell=45042
	// shared.NewSimpleStatActive(34430) // Glimmering Naaru Sliver - https://www.wowhead.com/tbc/spell=45052
	// shared.NewSimpleStatActive(34471) // Vial of the Sunwell - https://www.wowhead.com/tbc/spell=45064
	// shared.NewSimpleStatActive(35275) // Orb of the Sin'dorei - https://www.wowhead.com/tbc/spell=46354
	// shared.NewSimpleStatActive(35514) // Frostscythe of Lord Ahune - https://www.wowhead.com/tbc/spell=46643
	// shared.NewSimpleStatActive(35581) // Rocket Boots Xtreme Lite - https://www.wowhead.com/tbc/spell=51582
	// shared.NewSimpleStatActive(35703) // Figurine - Seaspray Albatross - https://www.wowhead.com/tbc/spell=46785
	// shared.NewSimpleStatActive(38175) // The Horseman's Blade - https://www.wowhead.com/tbc/spell=50070

	// Agility
	shared.NewSimpleStatActive(32658) // Badge of Tenacity - https://www.wowhead.com/tbc/spell=40729

	// Agility / Stamina / Strength
	shared.NewSimpleStatActive(15873) // Ragged John's Neverending Cup - https://www.wowhead.com/tbc/spell=20587

	// ArcaneResistance / FireResistance / FrostResistance / NatureResistance / ShadowResistance
	shared.NewSimpleStatActive(15867) // Prismcharm - https://www.wowhead.com/tbc/spell=19638
	shared.NewSimpleStatActive(23042) // Loatheb's Reflection - https://www.wowhead.com/tbc/spell=28778

	// Armor
	shared.NewSimpleStatActive(19345) // Aegis of Preservation - https://www.wowhead.com/tbc/spell=23780
	shared.NewSimpleStatActive(27891) // Adamantine Figurine - https://www.wowhead.com/tbc/spell=33479
	shared.NewSimpleStatActive(33830) // Ancient Aqir Artifact - https://www.wowhead.com/tbc/spell=43713

	// Armor / AttackPower / SpellDamage
	shared.NewSimpleStatActive(19337) // The Black Book - https://www.wowhead.com/tbc/spell=23720

	// ArmorPenetration
	shared.NewSimpleStatActive(28121) // Icon of Unyielding Courage - https://www.wowhead.com/tbc/spell=34106

	// AttackPower / RangedAttackPower
	shared.NewSimpleStatActive(14554) // Cloudkeeper Legplates - https://www.wowhead.com/tbc/spell=18787
	shared.NewSimpleStatActive(21180) // Earthstrike - https://www.wowhead.com/tbc/spell=25891
	shared.NewSimpleStatActive(23041) // Slayer's Crest - https://www.wowhead.com/tbc/spell=28777
	shared.NewSimpleStatActive(25628) // Ogre Mauler's Badge - https://www.wowhead.com/tbc/spell=32362
	shared.NewSimpleStatActive(25633) // Uniting Charm - https://www.wowhead.com/tbc/spell=32362
	shared.NewSimpleStatActive(25937) // Terokkar Tablet of Precision - https://www.wowhead.com/tbc/spell=39200
	shared.NewSimpleStatActive(25994) // Rune of Force - https://www.wowhead.com/tbc/spell=32955
	shared.NewSimpleStatActive(28041) // Bladefist's Breadth - https://www.wowhead.com/tbc/spell=33667
	shared.NewSimpleStatActive(29383) // Bloodlust Brooch - https://www.wowhead.com/tbc/spell=35166
	shared.NewSimpleStatActive(29776) // Core of Ar'kelos - https://www.wowhead.com/tbc/spell=35733
	shared.NewSimpleStatActive(30629) // Scarab of Displacement - https://www.wowhead.com/tbc/spell=38351
	shared.NewSimpleStatActive(31617) // Ancient Draenei War Talisman - https://www.wowhead.com/tbc/spell=33667
	shared.NewSimpleStatActive(32654) // Crystalforged Trinket - https://www.wowhead.com/tbc/spell=40724
	shared.NewSimpleStatActive(33831) // Berserker's Call - https://www.wowhead.com/tbc/spell=43716
	shared.NewSimpleStatActive(38287) // Empty Mug of Direbrew - https://www.wowhead.com/tbc/spell=51955

	// DodgeRating
	shared.NewSimpleStatActive(25787) // Charm of Alacrity - https://www.wowhead.com/tbc/spell=32600
	shared.NewSimpleStatActive(28528) // Moroes' Lucky Pocket Watch - https://www.wowhead.com/tbc/spell=34519

	// FireResistance
	shared.NewSimpleStatActive(13164) // Heart of the Scale - https://www.wowhead.com/tbc/spell=17275

	// HealingPower / SpellDamage
	shared.NewSimpleStatActive(18820) // Talisman of Ephemeral Power - https://www.wowhead.com/tbc/spell=23271
	shared.NewSimpleStatActive(19344) // Natural Alignment Crystal - https://www.wowhead.com/tbc/spell=23734
	shared.NewSimpleStatActive(20636) // Hibernation Crystal - https://www.wowhead.com/tbc/spell=24998
	shared.NewSimpleStatActive(22268) // Draconic Infused Emblem - https://www.wowhead.com/tbc/spell=27675
	shared.NewSimpleStatActive(22678) // Talisman of Ascendance - https://www.wowhead.com/tbc/spell=28200
	shared.NewSimpleStatActive(23046) // The Restrained Essence of Sapphiron - https://www.wowhead.com/tbc/spell=28779
	shared.NewSimpleStatActive(23047) // Eye of the Dead - https://www.wowhead.com/tbc/spell=28780
	shared.NewSimpleStatActive(25619) // Glowing Crystal Insignia - https://www.wowhead.com/tbc/spell=32355
	shared.NewSimpleStatActive(25620) // Ancient Crystal Talisman - https://www.wowhead.com/tbc/spell=32355
	shared.NewSimpleStatActive(25634) // Oshu'gun Relic - https://www.wowhead.com/tbc/spell=32367
	shared.NewSimpleStatActive(25936) // Terokkar Tablet of Vim - https://www.wowhead.com/tbc/spell=39201
	shared.NewSimpleStatActive(25995) // Star of Sha'naar - https://www.wowhead.com/tbc/spell=32956
	shared.NewSimpleStatActive(27828) // Warp-Scarab Brooch - https://www.wowhead.com/tbc/spell=33400
	shared.NewSimpleStatActive(28040) // Vengeance of the Illidari - https://www.wowhead.com/tbc/spell=33662
	shared.NewSimpleStatActive(28223) // Arcanist's Stone - https://www.wowhead.com/tbc/spell=34000
	shared.NewSimpleStatActive(29132) // Scryer's Bloodgem - https://www.wowhead.com/tbc/spell=35337
	shared.NewSimpleStatActive(29179) // Xi'ri's Gift - https://www.wowhead.com/tbc/spell=35337
	shared.NewSimpleStatActive(29370) // Icon of the Silver Crescent - https://www.wowhead.com/tbc/spell=35163
	shared.NewSimpleStatActive(29376) // Essence of the Martyr - https://www.wowhead.com/tbc/spell=35165
	shared.NewSimpleStatActive(30293) // Heavenly Inspiration - https://www.wowhead.com/tbc/spell=36347
	shared.NewSimpleStatActive(31615) // Ancient Draenei Arcane Relic - https://www.wowhead.com/tbc/spell=33662
	shared.NewSimpleStatActive(33828) // Tome of Diabolic Remedy - https://www.wowhead.com/tbc/spell=43710
	shared.NewSimpleStatActive(33829) // Hex Shrunken Head - https://www.wowhead.com/tbc/spell=43712
	shared.NewSimpleStatActive(38288) // Direbrew Hops - https://www.wowhead.com/tbc/spell=51954
	shared.NewSimpleStatActive(38290) // Dark Iron Smoking Pipe - https://www.wowhead.com/tbc/spell=51953

	// Health
	shared.NewSimpleStatActive(33832) // Battlemaster's Determination - https://www.wowhead.com/tbc/spell=44055
	shared.NewSimpleStatActive(34049) // Battlemaster's Audacity - https://www.wowhead.com/tbc/spell=44055
	shared.NewSimpleStatActive(34050) // Battlemaster's Perseverance - https://www.wowhead.com/tbc/spell=44055
	shared.NewSimpleStatActive(34162) // Battlemaster's Depravity - https://www.wowhead.com/tbc/spell=44055
	shared.NewSimpleStatActive(34163) // Battlemaster's Cruelty - https://www.wowhead.com/tbc/spell=44055
	shared.NewSimpleStatActive(35326) // Battlemaster's Alacrity - https://www.wowhead.com/tbc/spell=44055
	shared.NewSimpleStatActive(35327) // Battlemaster's Alacrity - https://www.wowhead.com/tbc/spell=44055

	// MeleeCritRating
	shared.NewSimpleStatActive(24114) // Braided Eternium Chain - https://www.wowhead.com/tbc/spell=31025

	// MeleeHasteRating
	shared.NewSimpleStatActive(22954) // Kiss of the Spider - https://www.wowhead.com/tbc/spell=28866
	shared.NewSimpleStatActive(28288) // Abacus of Violent Odds - https://www.wowhead.com/tbc/spell=33807

	// MeleeHasteRating / SpellHasteRating
	shared.NewSimpleStatActive(19343) // Scrolls of Blinding Light - https://www.wowhead.com/tbc/spell=23733

	// SpellCritRating
	shared.NewSimpleStatActive(19952) // Gri'lek's Charm of Valor - https://www.wowhead.com/tbc/spell=24498
	shared.NewSimpleStatActive(19957) // Hazza'rah's Charm of Destruction - https://www.wowhead.com/tbc/spell=24543

	// SpellDamage
	shared.NewSimpleStatActive(19959) // Hazza'rah's Charm of Magic - https://www.wowhead.com/tbc/spell=24544
	shared.NewSimpleStatActive(30340) // Starkiller's Bauble - https://www.wowhead.com/tbc/spell=36432

	// SpellDamage / SpellPenetration
	shared.NewSimpleStatActive(21473) // Eye of Moam - https://www.wowhead.com/tbc/spell=26166

	// SpellHasteRating
	shared.NewSimpleStatActive(19339) // Mind Quickening Gem - https://www.wowhead.com/tbc/spell=23723
	shared.NewSimpleStatActive(19955) // Wushoolay's Charm of Nature - https://www.wowhead.com/tbc/spell=24542
	shared.NewSimpleStatActive(19958) // Hazza'rah's Charm of Healing - https://www.wowhead.com/tbc/spell=24546
	shared.NewSimpleStatActive(32483) // The Skull of Gul'dan - https://www.wowhead.com/tbc/spell=40396

	// SpellHitRating
	shared.NewSimpleStatActive(19947) // Nat Pagle's Broken Reel - https://www.wowhead.com/tbc/spell=24610

	// Spirit
	shared.NewSimpleStatActive(28370) // Bangle of Endless Blessings - https://www.wowhead.com/tbc/spell=34210
	shared.NewSimpleStatActive(30665) // Earring of Soulful Meditation - https://www.wowhead.com/tbc/spell=40402

	// Strength
	shared.NewSimpleStatActive(28484) // Bulwark of Kings - https://www.wowhead.com/tbc/spell=34511
	shared.NewSimpleStatActive(28485) // Bulwark of the Ancient Kings - https://www.wowhead.com/tbc/spell=34511
}
