# Phyloformer training tracker

This is a simple interface to track training runs on remote servers. 

## Setting up
You can install the lastest version of [golang](https://go.dev/doc/install), clone this repo and build the binary with `go build`.

You can then run the executable and start adding training runs you want to track. 

You will need `rsync` installed on your system for this to work. 

## Tracking runs
You can add runs by clickign the `+` button on the top right of the screen. A random name will be assigned that will be used as a unique run identifier.  
You can then specify the fully qualified path to the csv containing the training and validation losses *(cf below)*, e.g. :  
`server-name:/path/to/the/tracking.csv`  
The program expects the csv you give it to have the following fields: `['train_loss', 'val_loss', 'step', 'lr']`, in any order you wish.  
By default it will update all running trainings every 5 minutes using rsync. In order to do this you need to setup you `~/.ssh/config` file in order to be able to automatically run rsync without prompting for a password. You can also force an update by clicking on the update button. 

## Developing
This repo uses [TailwindCSS](https://tailwindcss.com) in order to style the templates. In order to make changes you can install the [TailwindCSS CLI](https://tailwindcss.com/blog/standalone-cli) and run `./tailwindcss -i static/styles/tailwind.css -o static/styles/final.css -w -m` in the background from the root of the project. 

## TODO
- [ ] Add functional tags to organize and filter runs
- [ ] Add rolling average curve so you can see training trends. 
- [ ] ? 
