import 'package:flutter/material.dart';
import 'package:jim/src/constants/image_strings.dart';
import 'package:jim/src/constants/sizes.dart';
import 'package:jim/src/constants/text_strings.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:jim/src/screens/login_screen.dart';
import 'package:get/get.dart';
import 'package:jim/src/screens/register_screen.dart';

class WelcomeScreen extends StatelessWidget {
  const WelcomeScreen({super.key});

  // This widget is the root of your application.
  @override
  Widget build(BuildContext context) {
    var height = MediaQuery.of(context).size.height;

    return GetMaterialApp(
        debugShowCheckedModeBanner: false,
        home: Scaffold(
          body: Container(
            padding: EdgeInsets.all(tDefaultSize),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.spaceEvenly,
              children: [
                Text(AppName,
                    style: GoogleFonts.lilitaOne(fontSize: 40),
                    textAlign: TextAlign.center),
                Image(
                    image: AssetImage(WelcomeScreenImage),
                    height: height * 0.4),
                Column(
                  children: [
                    Text(Welcome,
                        style: GoogleFonts.anton(fontSize: 35),
                        textAlign: TextAlign.center),
                    SizedBox(height: 10),
                    Text(WelcomeMessage,
                        style: GoogleFonts.pacifico(
                            fontSize: 20, color: Colors.grey),
                        textAlign: TextAlign.center),
                  ],
                ),
                Row(
                  children: [
                    Expanded(
                      child: OutlinedButton(
                          onPressed: () => Get.to(()=> const LoginScreen()), style: OutlinedButton.styleFrom(shape: RoundedRectangleBorder()),child: Text(tLogin.toUpperCase(), style: TextStyle(color: Colors.black, fontSize: 20))),
                    ),
                    SizedBox(width: 10.0),
                    Expanded(
                      child: ElevatedButton(
                          onPressed: () => Get.to(()=> const RegisterScreen()), style: OutlinedButton.styleFrom(shape: RoundedRectangleBorder(), backgroundColor: Colors.black),child: Text(tsignup.toUpperCase(), style:TextStyle(color: Colors.white, fontSize: 20))),
                    ),
                  ],
                )
              ],
            ),
          ),
        ));
  }
}
