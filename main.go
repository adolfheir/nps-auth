package main

import "nps-auth/cmd"

func main() {

	// npsApi := npsapi.NewAPI("http://175.27.193.51:20100/", "ihouqi2022")

	// tunnelInfo, err := npsApi.GetOneTunnel(npsapi.GetOneTunnelReq{
	// 	ID: "1",
	// })

	// if err != nil {
	// 	log.Error().Err(err).Msg("failed to do request")
	// 	return
	// }
	// log.Info().Msgf("tunnel info: %+v", tunnelInfo.AjaxOne.Data)
	// log.Info().Interface("tunnelInfo", tunnelInfo).Msg("get tunnel info")

	// os := runtime.GOOS
	// arch := runtime.GOARCH

	// println(os, arch)

	cmd.Execute()

}
