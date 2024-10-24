import 'package:flutter/material.dart';
import 'package:jim/src/constants/sizes.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:get/get.dart';
import 'package:jim/src/screens/otp_screen.dart';
import 'base_client.dart';
import 'package:flutter_otp_text_field/flutter_otp_text_field.dart';

class EmailVerfication extends StatefulWidget {
  const EmailVerfication({Key? key}) : super(key: key);
  @override
  _EmailVerification createState() => _EmailVerification();
}
class _EmailVerification extends State<EmailVerfication> {
  @override
  Widget build(BuildContext context){
    return SafeArea(
      child: Scaffold(
        body: SingleChildScrollView(
          child: Container(
            padding: const EdgeInsets.all(tDefaultSize),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,

              children: [
                SizedBox(height: tDefaultSize),
                Text(
                  'CO\nDE',
                  style: GoogleFonts.montserrat(fontSize: 80, fontWeight: FontWeight.bold),
                ),
                Text(
                    'VERIFICATION',
                    style: GoogleFonts.anton(fontSize: 30)
                ),
                SizedBox(height: 20),
                Text(
                  'Enter the verification code sent to your\nemail.',
                  style: GoogleFonts.cormorant(fontSize: 20),textAlign: TextAlign.center,
                ),
                SizedBox(height: 20),

                OtpTextField(
                  mainAxisAlignment: MainAxisAlignment.center,
                  numberOfFields: 6,
                  fillColor: Colors.black.withOpacity(0.1),
                  filled: true,
                  onSubmit: (code){print("OTP is: => $code");},
                ),
                SizedBox(height: 20),
                SizedBox(
                  width: double.infinity,
                  child: ElevatedButton(
                    onPressed: () {
                      // Your onPressed code here
                    },
                    style: ElevatedButton.styleFrom(
                      backgroundColor: Colors.black, // Set the background color to black
                    ),
                    child: Text(
                      "NEXT",
                      style: TextStyle(color: Colors.white), // Optionally set the text color to white for contrast
                    ),
                  ),
                )


              ],
            ),
          ),
        ),
      ),
    );
  }
}