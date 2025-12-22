package utils

import "fmt"

const logo = `
    ____             _      __               
   / __ \____  _____(_)____/ /_  ____  __  __
  / /_/ / __ \/ ___/ / ___/ __ \/ __ \/ / / /
 / ____/ /_/ / /  / / /__/ / / / /_/ / /_/ / 
/_/    \____/_/  /_/\___/_/ /_/\____/\__, /  
                                    /____/   

`

func Logo() string {
	return logo
}

func Welcome() {
	fmt.Println(logo)
}
