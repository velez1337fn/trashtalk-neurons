#!/usr/bin/env bun

const CYAN = "\x1b[36m"
const GREEN = "\x1b[32m"
const YELLOW = "\x1b[33m"
const BLUE = "\x1b[34m"
const RED = "\x1b[31m"
const MUTED = "\x1b[0;2m"
const NC = "\x1b[0m"
const WHITE = "\x1b[37m"

const LOGO = `${CYAN}   .___                 __         .__  .__                              ____    _______      _______
   |   | ____   _______/  |______  |  | |  |   ___________      ___  __ /_   |   \\   _  \\     \\   _  \\
   |   |/    \\ /  ___/\\   __\\__  \\ |  | |  | _/ __ \\_  __ \\     \\  \\/ /  |   |   /  /_\\  \\    /  /_\\  \\
   |   |   |  \\\\___ \\  |  |  / __ \\|  |_|  |\\_  ___/|  | \\/      \\   /   |   |   \\  \\_/   \\   \\  \\_/   \\
   |___|___|  /____  > |__| (____  /____/____/\\___  >__|          \\_/    |___| /\\ \\_____  / /\\ \\_____  /
            \\/     \\/            \\/               \\/                           \\/       \\/  \\/       \\/${NC}`

const { env } = require("process")
const { mkdir, rename, chmod, symlink, open, read, write } = require("fs/promises")
const { execSync } = require("child_process")
const { join } = require("path")

async function main() {
  const user = env.USER || env.HOME?.split("/").pop() || "user"

  let arch = "x86_64"
  try { arch = execSync("uname -m").toString().trim() } catch {}
  const displayArch = arch === "x86_64" ? "amd64" : "arm64"
  const sysInfo = `linux-${displayArch}`

  console.log(LOGO)
  console.log("")
  console.log(`${GREEN}┌─────────────────────┐${NC}${" ".repeat(56)}${GREEN}┌──────────────────────────────────────────┐${NC}`)
  console.log(`${GREEN}│${NC} ${CYAN}welcome, ${user}!${NC}       ${GREEN}│${NC}${" ".repeat(56)}${GREEN}│${NC} ${" ".repeat(58)}${GREEN}│${NC}`)
  console.log(`${GREEN}│${NC}${GREEN}├─────────────────────┤${NC}${" ".repeat(56)}${GREEN}├──────────────────────────────────────────┤${NC}`)
  console.log(`${GREEN}│${NC} ${BLUE}matrix in this box${NC} ${GREEN}│${NC}${" ".repeat(56)}${GREEN}│${NC} ${" ".repeat(58)}${GREEN}│${NC}`)
  console.log(`${GREEN}│${NC}${GREEN}├─────────────────────┐${NC}${" ".repeat(56)}${GREEN}├──────────────────────────────────────────┤${NC}`)
  console.log(`${GREEN}│${NC} ${YELLOW}detected sys:${NC} ${sysInfo}${GREEN}│${NC}${" ".repeat(56)}${GREEN}│${NC} ${" ".repeat(58)}${GREEN}│${NC}`)
  console.log(`${GREEN}│${NC} ${YELLOW}installer status:${NC} ready${GREEN}│${NC}${" ".repeat(56)}${GREEN}│${NC} ${" ".repeat(58)}${GREEN}│${NC}`)
  console.log(`${GREEN}│${NC} ${YELLOW}waiting for user input...${NC}${GREEN}│${NC}${" ".repeat(56)}${GREEN}│${NC} ${RED}are ya winning son?     ***   ***${NC}     ${GREEN}│${NC}`)
  console.log(`${GREEN}│${NC}                     ${GREEN}│${NC}${" ".repeat(56)}${GREEN}│${NC} ${RED}    ─┬────┐${NC}             ***** *****       ${GREEN}│${NC}`)
  console.log(`${GREEN}│${NC}                     ${GREEN}│${NC}${" ".repeat(56)}${GREEN}│${NC} ${RED}     └─┬──┘${NC}              *********        ${GREEN}│${NC}`)
  console.log(`${GREEN}│${NC}                     ${GREEN}│${NC}${" ".repeat(56)}${GREEN}│${NC} ${RED}   ┌───┼──┐${NC}               *******         ${GREEN}│${NC}`)
  console.log(`${GREEN}│${NC}                     ${GREEN}│${NC}${" ".repeat(56)}${GREEN}│${NC} ${RED}   │   │  │${NC}                *****          ${GREEN}│${NC}`)
  console.log(`${GREEN}│${NC}                     ${GREEN}│${NC}${" ".repeat(56)}${GREEN}│${NC} ${RED}     ┌─┼─┐${NC}                  ***           ${GREEN}│${NC}`)
  console.log(`${GREEN}│${NC}                     ${GREEN}│${NC}${" ".repeat(56)}${GREEN}│${NC} ${RED}     │   │${NC}                   *            ${GREEN}│${NC}`)
  console.log(`${GREEN}│${NC}                     ${GREEN}│${NC}${" ".repeat(56)}${GREEN}│${NC} ${RED}     └─┬─┬┘${NC}                    ──           ${GREEN}│${NC}`)
  console.log(`${GREEN}│${NC}                     ${GREEN}│${NC}${" ".repeat(56)}${GREEN}│${NC} ${RED}       └─┘${NC}                         ***        ${GREEN}│${NC}`)
  console.log(`${GREEN}└─────────────────────┴─────────────────────────────────────────────────────────┴─────┴───┴────────────────────────────────┘${NC}`)
  console.log("")
  console.log(`  ${GREEN}1.${NC} Install TrashNeurons CLI`)
  console.log(`  ${GREEN}2.${NC} Exit`)
  console.log("")

  const choice = await promptChoice()

  if (choice === 2) {
    console.log("Goodbye!")
    return
  }
  if (choice !== 1) {
    console.log(`${RED}Invalid option.${NC}`)
    return
  }

  await install(displayArch)
}

function promptChoice(): Promise<number> {
  return new Promise((resolve) => {
    const readline = require("readline").createInterface({ input: process.stdin, output: process.stdout })
    readline.question("Choose option [1-2]: ", (answer) => {
      readline.close()
      const num = parseInt(answer.trim())
      if (num === 1 || num === 2) resolve(num)
      else resolve(0)
    })
  })
}

async function install(targetArch) {
  console.log("")
  console.log(`${BLUE}┌────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐${NC}`)
  console.log(`${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}`)

  console.log(`${BLUE}│${NC} ${YELLOW}>${NC} allocating path where app/cli been installed...                                                                    ${BLUE}│${NC}`)
  console.log(`${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}`)

  const home = env.HOME
  const installDir = join(home, ".trashneurons")
  const binDir = join(installDir, "bin")
  await mkdir(binDir, { recursive: true })

  console.log(`${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}`)
  console.log(`${BLUE}│${NC}                 ${GREEN}found! path is:${NC} ${installDir}                                                                            ${BLUE}│${NC}`)
  console.log(`${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}`)
  console.log(`${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}`)
  console.log(`${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}`)
  console.log(`${BLUE}│${NC} ${YELLOW}>${NC} starting download via curl...                                                                                       ${BLUE}│${NC}`)
  console.log(`${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}`)

  const resp = await fetch("https://api.github.com/repos/velez1337fn/trashtalk-neurons/releases/latest")
  const release = await resp.json()

  const asset = release.assets.find(a => a.name.includes(targetArch))

  if (!asset) {
    console.log(`${BLUE}│${NC} ${RED}ERROR: No release found for arch ${targetArch}${NC}                                                                               ${BLUE}│${NC}`)
    console.log(`${BLUE}│${NC} Please run the GitHub Actions workflow first to build the binary.                                                      ${BLUE}│${NC}`)
    console.log(`${BLUE}└────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘${NC}`)
    return
  }

  console.log(`${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}`)
  console.log(`${BLUE}│${NC} ${GREEN}found release! downloading...${NC}                                                                                         ${BLUE}│${NC}`)
  console.log(`${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}`)

  const dlResp = await fetch(asset.browser_download_url)
  const tmpPath = `/tmp/trashneurons-installer-bin-${targetArch}`
  const file = await open(tmpPath, "w")
  await dlResp.body.pipe(file)
  file.close()
  await chmod(tmpPath, 0o755)

  console.log(`${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}`)
  console.log(`${BLUE}│${NC} ${GREEN}downloaded!${NC}                                                                                                           ${BLUE}│${NC}`)
  console.log(`${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}`)
  console.log(`${BLUE}│${NC}    ${YELLOW}installing your app/cli...${NC}                                                                                          ${BLUE}│${NC}`)
  console.log(`${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}`)

  const destBin = join(binDir, "trashneurons-cli")
  await rename(tmpPath, destBin)

  try {
    await symlink(destBin, "/usr/local/bin/trashneurons-cli")
  } catch {
    // need sudo, skip
  }

  console.log(`${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}`)
  console.log(`${BLUE}│${NC}  ${GREEN}finished install!${NC} all files located in ${installDir}                                                                   ${BLUE}│${NC}`)
  console.log(`${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}`)
  console.log(`${BLUE}│${NC}  ${YELLOW}for use:${NC} type "trashneurons-cli"                                                                                    ${BLUE}│${NC}`)
  console.log(`${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}`)
  console.log(`${BLUE}├──────────────────────┐${NC}${" ".repeat(81)}${BLUE}├────────────────────────────────────────────────────────┤${NC}`)
  console.log(`${BLUE}│${NC} ${RED}PRESS CTRL+C TO CLOSE INSTALLER${NC}   ${BLUE}│${NC}${" ".repeat(81)}${BLUE}│${NC}`)
  console.log(`${BLUE}└──────────────────────┴─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘${NC}`)
  console.log("")
  console.log(`${GREEN}Press Enter to exit...${NC}`)
  await new Promise((resolve) => {
    const readline = require("readline").createInterface({ input: process.stdin, output: process.stdout })
    readline.once("close", resolve)
  })
}

main()
