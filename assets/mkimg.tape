# Create the terminal demonstration:
#
#   go install github.com/charmbracelet/vhs@latest
#   sudo snap install ttyd --classic
#   vhs mkimg.tape
#
# Then compress with gif2webp:
#
#   gif2webp -mixed -m 6 -q 50 -kmin 0 -kmax 0 -min_size mkimg.gif -o mkimg.webp
#
# -----------------------------------------------------------------------------

Output mkimg.gif

# -----------------------------------------------------------------------------
# Settings
# -----------------------------------------------------------------------------

Set FontSize 24
Set Framerate 24
Set Height 640
Set Padding 36
Set Width 1300
Set TypingSpeed 50ms

# -----------------------------------------------------------------------------
# Setup
# -----------------------------------------------------------------------------

Hide
Type "source ~/.bashrc"
Enter
Type "PS1='\033[33m$\033[0m '"
Enter
Ctrl+L
Show

# -----------------------------------------------------------------------------
# Examples
# -----------------------------------------------------------------------------

Type "mkimg -o blue.jpg 6 7 blue"
Sleep 1
Enter
Wait+Line@10s /\$/ # wait for prompt to appear
Sleep 2

Type "mkimg -o blue.jpg 6 7 '#0000ff'"
Sleep 1
Enter
Wait+Line@10s /\$/ # wait for prompt to appear
Sleep 2

Type "mkimg -o blue.jpg 6 7 '#00f'"
Sleep 1
Enter
Wait+Line@10s /\$/ # wait for prompt to appear
Sleep 2

Type "mkimg -o blue.jpg 6 7 0000ff"
Sleep 1
Enter
Wait+Line@10s /\$/ # wait for prompt to appear
Sleep 2

Type "mkimg -o blue.jpg 6 7 00f"
Sleep 1
Enter
Wait+Line@10s /\$/ # wait for prompt to appear
Sleep 2

Type "mkimg -o semi-transparent-red.png 42 42 '#ff000088'"
Sleep 1
Enter
Wait+Line@10s /\$/ # wait for prompt to appear
Sleep 2

Type "mkimg -o semi-transparent-red.png 42 42 '#f008'"
Sleep 1
Enter
Wait+Line@10s /\$/ # wait for prompt to appear
Sleep 2

Type "mkimg -o semi-transparent-red.png 42 42 ff000088"
Sleep 1
Enter
Wait+Line@10s /\$/ # wait for prompt to appear
Sleep 2

Type "mkimg -o semi-transparent-red.png 42 42 f008"
Sleep 1
Enter
Wait+Line@10s /\$/ # wait for prompt to appear
Sleep 2

Type "mkimg -o pixel.png 1 1 transparent"
Sleep 1
Enter
Wait+Line@10s /\$/ # wait for prompt to appear
Sleep 2

Type "mkimg -f png -b 1 1 transparent"
Sleep 1
Enter
Wait+Line@10s /\$/ # wait for prompt to appear
Sleep 2

Type "mkimg -b 1 1 transparent"
Sleep 1
Enter
Wait+Line@10s /\$/ # wait for prompt to appear
Sleep 2

# -----------------------------------------------------------------------------
# Wait a few more seconds before next loop
# -----------------------------------------------------------------------------

Sleep 2
