#!/bin/bash

set -e

readonly FCTL_URL='https://github.com/TTraveller7/invokerlib/releases/download/v0.6.0/fctl'
readonly DEFAULT_INSTALL_PATH="/opt/$USER"

# Step 1: Create directory
read -p "Enter directory to place fctl (default: $DEFAULT_INSTALL_PATH): " install_path
if [ -z "$install_path" ]
then 
    install_path=$DEFAULT_INSTALL_PATH
fi
# append fctl to the install path
install_path="$install_path/fctl"
if [ -d "$install_path" ] 
then 
    rm -rf $install_path
fi
echo "fctl will be installed at: $install_path"
mkdir -p $install_path

# Step 2: Fetch fctl executable
echo "Fetching fctl to directory $install_path"
wget -q -P $install_path $FCTL_URL
fctlPath="$install_path/fctl"
chmod +x $fctlPath
echo "Successfully fectched fctl to $fctlPath"

# Step 3: Add install path to bash profile
readonly BASH_PRE_STR='export PATH=$PATH'
readonly BASH_PROFILE_LINE="$BASH_PRE_STR:$install_path"
readonly BASH_PROFILE_PATH="$HOME/.bash_profile"
if ! grep -q "$BASH_PROFILE_LINE" $BASH_PROFILE_PATH
then 
    echo "Adding the following line to $BASH_PROFILE_PATH: $BASH_PROFILE_LINE"
    echo $BASH_PROFILE_LINE >> $BASH_PROFILE_PATH
fi

echo 'fctl is successfully installed. Restart your terminal or use `source .bash_profile` to update PATH variable.'
echo 'After that, Try `fctl init`.'
