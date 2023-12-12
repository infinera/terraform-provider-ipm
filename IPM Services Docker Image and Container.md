# Build IPM Services Docker Image Steps
1. Preconditions: GoLang and docker are installed.
2. Clone [Terraform IPM provider](https://bitbucket.infinera.com/projects/MAR/repos/terraform-provider-ipm/browse) in to working directory
3. In the [Terraform IPM provider](https://bitbucket.infinera.com/projects/MAR/repos/terraform-provider-ipm/browse) directory, issue command "make build/docker". After successful completion, a docker image with tag "sv-artifactory.infinera.com/marvel/ipm/ipm-services:v0.0.1" are available to be used.

# Run the IPM Services Docker Container
tainer using the locally built IPM Services Docker Image or from the Docker Repository Registry 
In any user working directory, execute the following command **"docker run -it -v $HostDir:/Work-Directory --add-host=$IPM_HOST_NAME:$IPM_HOST_IP_ADDRESS $IMAGE_NAME bash"**
1. **-it** the container will start in interactive mode where the user can execute command in the container shell.
2. **-v $HostDir:/Work-Directory** - specify container volume "/Work-Directory" which is mapped from the host "$HostDir" directory. All worked directories shall be created in this volume. This allow the user to add, update the intent files and user profiles from host system without the needs to build a new IPM Services Docker Image with the desired intents.
3. **--add-host=$IPM_HOST_NAME:\$IPM_HOST_IP_ADDRESS** - The IPM server host name and its IP address will be added to the /etc/hosts to allow accessing from the container.
4. **$IMAGE_NAME** - The image tag name and its version. The tag name and version is specified in the GNUmakefile:build/docker make rule which is *sv-artifactory.infinera.com/marvel/ipm/ipm-services:v0.0.1* currently 
5. **bash** the container shall start in a bash shell.

**Examples:** *docker run -it -v "$(pwd)":/Work-Directory --add-host=pt-xrivk824-dv:10.46.76.81 sv-artifactory.infinera.com/marvel/ipm/ipm-services:v0.0.1 bash*

# Manage the XR Network via Configuration Intent 
Now in the Container bash shell, the user can apply the desired intents to the network via different services to manage the XR network. 
## The Intent files
The user specifies the intent via the intent file in the */Work-Directory/$managed-resource/user-intents* directory. Notice that the intent files are from the host volume; hence the user can view, add, delete, update and save them in the host system directly. Any changes to the host volume intent files shall be available to the IPM service container immediately and vice versa. The *$managed-resource* directory which will be created as needed when the user execute the *setup.sh* command in the running *ipm-service* container. Please see next section for more description.
## Manage Networks
1. Bring up the docker container in any directory.
   1. *docker run -it -v "$(pwd)":/Work-Directory --add-host=pt-xrivk824-dv:10.46.76.81 sv-artifactory.infinera.com/marvel/ipm/ipm-services:v0.0.1 startdirectory*
   The *startdirectory* directory will be created as needed. The user can copy and/or create new intent files in the *user-intents* subdirectories and use them to create/update/delete the network. The user profile must be *network-profiles.json" for network resources, nc-profiles for Network Connection resources,etc. The setup script shall populate the user-profiles subdirectory with the right sample user profile and the it can be customized as required. Now The current directory is */Work-Directory/network1/*.
2. To view all modules in the XR Networks. Run the command *get_-_modules.sh*. If the execution is success, it shall generate *get-modules-output.json* file at */Work-Directory/startDirectory*.
3. To create a Constellation Network. Run the command *networks $1 init=yes intent=networks.tfvars*. 
   1. **Notice The option *init=yes* only require the first time execution**. 
   2. The *networks.tfvars* intent file must existed in the */Work-Directory/network1/user-intents* directory. 
   3. The support *$1* commands are
      1. create: If the execution is success, it shall generate *networks-output.json* file at */Work-Directory/network1*.
      2. update: If the execution is success, it shall generate *networks-output.json* file at */Work-Directory/network1*. Just update the intent file to update.
      3. plan: If the execution is success, it shall generate *networks-plan.json* file at */Work-Directory/network1*.
      4. delete: 

## Manage Transport Capacities - TBD - Not Ready Yet

## Manage Network Connections (NC) - TBD - Not Ready Yet
   
## Trouble Shooting
1. If terraform **apply** fails for any reason, the terraform state file may need to be removed and run "terraform init" again. Execute the command "*networks $1 init=yes intent=networks.tfvars*
